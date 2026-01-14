CREATE TABLE IF NOT EXISTS task_statuses (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(30) NOT NULL,  -- 'todo', 'in_progress', 'done', 'archived'
  description  VARCHAR(50) NOT NULL,
  display_order INTEGER NOT NULL,     -- UI表示順序 (例: 1, 2, 3, 4)
  is_active BOOLEAN NOT NULL DEFAULT true,  -- 論理削除フラグ
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP DEFAULT NULL
);

CREATE UNIQUE INDEX idx_task_statuses_name ON task_statuses(name);
CREATE INDEX idx_task_statuses_display_order ON task_statuses(display_order);

-- 初期データ
INSERT INTO task_statuses (name, description, display_order) VALUES
  ('todo', '未着手', 1),
  ('in_progress', '進行中', 2),
  ('done', '完了', 3),
  ('archived', 'アーカイブ', 4);
