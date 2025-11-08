BEGIN

DROP TABLE IF EXISTS tasks;

CREATE TABLE IF NOT EXISTS tasks (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  title VARCHAR(200) NOT NULL,
  description TEXT,
  status_id BIGINT NOT NULL DEFAULT 1, -- task_statuses.idへの外部キー
  priority VARCHAR(30) NOT NULL DEFAULT 'medium',
  due_date TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  
  CONSTRAINT fk_tasks_user_id 
    FOREIGN KEY (user_id) 
    REFERENCES users(id) 
    ON DELETE CASCADE,

  CONSTRAINT fk_tasks_status_id 
    FOREIGN KEY (status_id) 
    REFERENCES task_statuses(id) 
    ON DELETE RESTRICT
);

CREATE INDEX idx_tasks_status_id ON tasks(status_id);
CREATE INDEX idx_tasks_user_id ON tasks(user_id);
CREATE INDEX idx_tasks_due_date ON tasks(due_date);

COMMIT;

-- 変更を取り消したい場合に使用する。
-- ROLLBACK;
