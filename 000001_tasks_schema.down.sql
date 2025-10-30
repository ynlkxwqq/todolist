-- DROP TABLE IF EXISTS tasks;
-- migrations/000001_create_tasks.up.sql
CREATE TABLE IF NOT EXISTS tasks (
  id UUID PRIMARY KEY,
  title VARCHAR(200) NOT NULL,
  active_at DATE NOT NULL,
  done BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- уникальность по title + activeAt
CREATE UNIQUE INDEX IF NOT EXISTS idx_tasks_title_activeat ON tasks (title, active_at);
