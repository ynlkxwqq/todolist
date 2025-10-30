// package main

// import (
// 	"embed"
// 	"io/fs"
// 	"todo-list/internal/app"
// 	"todo-list/internal/server"
// )

// //go:embed api/docs/*
// var dist embed.FS

// // FS holds embedded swagger-ui files
// var FS, _ = fs.Sub(dist, "api/docs")

// func main() {
// 	server.FS = FS

// 	app.Run()
// }

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
	_ "github.com/lib/pq"
)

type Task struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	ActiveAt string    `json:"activeAt"` // YYYY-MM-DD
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
	dsn := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/todo?sslmode=disable")
	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer db.Close()

	// простая проверка соединения
	if err = db.Ping(); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	http.HandleFunc("/api/todo-list/tasks", tasksHandler)
	http.HandleFunc("/api/todo-list/tasks/", taskByIDHandler) // for paths with ID

	port := getEnv("PORT", "8080")
	log.Printf("server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

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

// POST /api/todo-list/tasks
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
	q := `INSERT INTO tasks(id, title, active_at) VALUES ($1,$2,$3)`
	_, err = db.Exec(q, id, req.Title, activeAt)
	if err != nil {
		// check unique constraint
		if isUniqueViolation(err) {
			httpError(w, http.StatusNotFound, "task already exists")
			return
		}
		httpError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusCreated)
	resp := map[string]string{"id": id}
	_ = json.NewEncoder(w).Encode(resp)
}

// PUT /api/todo-list/tasks/{ID}
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

	// обновляем
	result, err := db.Exec(`UPDATE tasks SET title=$1, active_at=$2 WHERE id=$3`, req.Title, activeAt, id)
	if err != nil {
		if isUniqueViolation(err) {
			httpError(w, http.StatusNotFound, "task already exists with same title and date")
			return
		}
		httpError(w, http.StatusInternalServerError, "internal error")
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		httpError(w, http.StatusNotFound, "task not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DELETE /api/todo-list/tasks/{ID}
func handleDeleteTask(w http.ResponseWriter, r *http.Request, id string) {
	if id == "" {
		httpError(w, http.StatusBadRequest, "id required")
		return
	}
	res, err := db.Exec(`DELETE FROM tasks WHERE id=$1`, id)
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

// PUT /api/todo-list/tasks/{ID}/done
func handleMarkDone(w http.ResponseWriter, r *http.Request, id string) {
	if id == "" {
		httpError(w, http.StatusBadRequest, "id required")
		return
	}
	res, err := db.Exec(`UPDATE tasks SET done = true WHERE id=$1`, id)
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

// GET /api/todo-list/tasks?status=active|done
func handleListTasks(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "active"
	}
	var rows *sql.Rows
	var err error
	today := time.Now().UTC().Format("2006-01-02")
	if status == "done" {
		rows, err = db.Query(`SELECT id, title, active_at, done, created_at FROM tasks WHERE done = true ORDER BY created_at`)
	} else {
		// active: active_at <= today and done = false
		rows, err = db.Query(`SELECT id, title, active_at, done, created_at FROM tasks WHERE done = false AND active_at <= $1 ORDER BY created_at`, today)
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
		// если выходной — добавляем префикс в возвращаемом title
		if isWeekend {
			t.Title = "ВЫХОДНОЙ - " + t.Title
		}
		tasks = append(tasks, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// общий роутинг по пути /api/todo-list/tasks/{...}
func taskByIDHandler(w http.ResponseWriter, r *http.Request) {
	// ожидаем пути вида /api/todo-list/tasks/{id} или /api/todo-list/tasks/{id}/done
	path := r.URL.Path
	// отрезаем префикс
	prefix := "/api/todo-list/tasks/"
	if len(path) <= len(prefix) {
		httpError(w, http.StatusNotFound, "not found")
		return
	}
	rest := path[len(prefix):] // {id} или {id}/done
	// если содержит "/done"
	if r.Method == http.MethodPut && len(rest) > 5 && rest[len(rest)-5:] == "/done" {
		id := rest[:len(rest)-5]
		handleMarkDone(w, r, id)
		return
	}

	// иначе rest == id
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

// ===== helpers =====
func httpError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func parseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, errors.New("empty date")
	}
	t, err := time.Parse("2006-01-02", s)
	return t, err
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
	// very simple check for pq unique violation text
	if err == nil {
		return false
	}
	return (err.Error() != "" && (contains(err.Error(), "unique") || contains(err.Error(), "duplicate key")))
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && ((len(s) > 0) && (stringIndex(s, sub) >= 0))
}

// simple string index (avoid importing strings lib multiple times)
func stringIndex(s, sep string) int {
	for i := 0; i+len(sep) <= len(s); i++ {
		if s[i:i+len(sep)] == sep {
			return i
		}
	}
	return -1
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
