package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type Task struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	ActiveAt string    `json:"activeAt"`
	Done     bool      `json:"done,omitempty"`
	Created  time.Time `json:"-"`
}

type CreateTaskRequest struct {
	Title    string `json:"title"`
	ActiveAt string `json:"activeAt"`
}

type UpdateTaskRequest struct {
	Title    string `json:"title"`
	ActiveAt string `json:"activeAt"`
}

var db *sql.DB

func main() {
	// === 1. Connect to SQLite ===
	dsn := getEnv("DATABASE_URL", "./todo.db")
	var err error
	db, err = sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// === 2. Create table if not exists ===
	schema := `
	CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		active_at DATE NOT NULL,
		done BOOLEAN DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(title, active_at)
	);`
	if _, err := db.Exec(schema); err != nil {
		log.Fatalf("create schema: %v", err)
	}

	http.HandleFunc("/api/todo-list/tasks", tasksHandler)
	http.HandleFunc("/api/todo-list/tasks/", taskByIDHandler)

	port := getEnv("PORT", "8080")
	log.Printf("✅ Server running on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// === HANDLERS ===

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handleCreateTask(w, r)
	case http.MethodGet:
		handleListTasks(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := validateTitle(req.Title); err != nil {
		httpError(w, http.StatusBadRequest, err.Error())
		return
	}
	activeAt, err := parseDate(req.ActiveAt)
	if err != nil {
		httpError(w, http.StatusBadRequest, "invalid activeAt date, expected YYYY-MM-DD")
		return
	}

	id := uuid.New().String()
	q := `INSERT INTO tasks(id, title, active_at) VALUES (?, ?, ?)`
	_, err = db.Exec(q, id, req.Title, activeAt)
	if err != nil {
		if isUniqueViolation(err) {
			httpError(w, http.StatusNotFound, "task already exists")
			return
		}
		httpError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func handleUpdateTask(w http.ResponseWriter, r *http.Request, id string) {
	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if id == "" {
		httpError(w, http.StatusBadRequest, "id required")
		return
	}
	if err := validateTitle(req.Title); err != nil {
		httpError(w, http.StatusBadRequest, err.Error())
		return
	}
	activeAt, err := parseDate(req.ActiveAt)
	if err != nil {
		httpError(w, http.StatusBadRequest, "invalid activeAt date, expected YYYY-MM-DD")
		return
	}

	res, err := db.Exec(`UPDATE tasks SET title=?, active_at=? WHERE id=?`, req.Title, activeAt, id)
	if err != nil {
		if isUniqueViolation(err) {
			httpError(w, http.StatusNotFound, "duplicate task")
			return
		}
		httpError(w, http.StatusInternalServerError, "internal error")
		return
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		httpError(w, http.StatusNotFound, "task not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleDeleteTask(w http.ResponseWriter, r *http.Request, id string) {
	if id == "" {
		httpError(w, http.StatusBadRequest, "id required")
		return
	}
	res, err := db.Exec(`DELETE FROM tasks WHERE id=?`, id)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "internal error")
		return
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		httpError(w, http.StatusNotFound, "task not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleMarkDone(w http.ResponseWriter, r *http.Request, id string) {
	if id == "" {
		httpError(w, http.StatusBadRequest, "id required")
		return
	}
	res, err := db.Exec(`UPDATE tasks SET done = 1 WHERE id=?`, id)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "internal error")
		return
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		httpError(w, http.StatusNotFound, "task not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleListTasks(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "active"
	}
	var rows *sql.Rows
	var err error
	today := time.Now().UTC().Format("2006-01-02")

	if status == "done" {
		rows, err = db.Query(`SELECT id, title, active_at, done, created_at FROM tasks WHERE done = 1 ORDER BY created_at`)
	} else {
		rows, err = db.Query(`SELECT id, title, active_at, done, created_at FROM tasks WHERE done = 0 AND active_at <= ? ORDER BY created_at`, today)
	}
	if err != nil {
		httpError(w, http.StatusInternalServerError, "internal error")
		return
	}
	defer rows.Close()

	tasks := []Task{}
	isWeekend := isWeekendDay(time.Now())
	for rows.Next() {
		var t Task
		var activeAt time.Time
		if err := rows.Scan(&t.ID, &t.Title, &activeAt, &t.Done, &t.Created); err != nil {
			httpError(w, http.StatusInternalServerError, "internal error")
			return
		}
		t.ActiveAt = activeAt.Format("2006-01-02")
		if isWeekend {
			t.Title = "ВЫХОДНОЙ - " + t.Title
		}
		tasks = append(tasks, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// === ROUTER ===
func taskByIDHandler(w http.ResponseWriter, r *http.Request) {
	prefix := "/api/todo-list/tasks/"
	if len(r.URL.Path) <= len(prefix) {
		httpError(w, http.StatusNotFound, "not found")
		return
	}
	rest := r.URL.Path[len(prefix):]

	if r.Method == http.MethodPut && len(rest) > 5 && rest[len(rest)-5:] == "/done" {
		id := rest[:len(rest)-5]
		handleMarkDone(w, r, id)
		return
	}

	id := rest
	switch r.Method {
	case http.MethodPut:
		handleUpdateTask(w, r, id)
	case http.MethodDelete:
		handleDeleteTask(w, r, id)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// === HELPERS ===
func httpError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func parseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, errors.New("empty date")
	}
	return time.Parse("2006-01-02", s)
}

func validateTitle(t string) error {
	if t == "" {
		return errors.New("title is required")
	}
	if len(t) > 200 {
		return errors.New("title must be <= 200 chars")
	}
	return nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return (err.Error() != "" && (contains(err.Error(), "UNIQUE") || contains(err.Error(), "constraint failed")))
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func isWeekendDay(t time.Time) bool {
	wd := t.Weekday()
	return wd == time.Saturday || wd == time.Sunday
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
