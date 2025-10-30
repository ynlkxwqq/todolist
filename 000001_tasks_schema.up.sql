-- CREATE TABLE IF NOT EXISTS tasks (
-- 	id VARCHAR(24) PRIMARY KEY,
-- 	title VARCHAR(200) UNIQUE NOT NULL,
-- 	active_at DATE UNIQUE,
-- 	status VARCHAR(20) DEFAULT 'active'
-- );

-- CREATE INDEX idx_tasks_title ON tasks (title);
-- CREATE INDEX idx_tasks_active_at ON tasks (active_at);