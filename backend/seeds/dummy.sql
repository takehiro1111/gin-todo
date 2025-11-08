
-- 外部キー制約の依存関係でマスターデータから挿入する。
INSERT INTO task_statuses (name, description, display_order) VALUES
  ('todo', '未着手のタスク', 1),
  ('in_progress', '作業中のタスク', 2),
  ('done', '完了したタスク', 3),
  ('archived', 'アーカイブされたタスク', 4);

INSERT INTO user_roles (name, display_order, is_active) VALUES
  ('admin', 1, true),
  ('writer', 2, true),
  ('viewer', 3, true);

INSERT INTO users (email, password_hash, name, role_id) VALUES
  ('admin@example.com', '$hashed_password', 'Admin User', 1),
  ('user1@example.com', '$hashed_password', 'Test User 1', 2),
  ('user2@example.com', '$hashed_password', 'Test User 2', 3);

INSERT INTO tasks (user_id, title, status_id, priority, due_date) VALUES
  (2, 'タスク1', 1, 'high', '2025-12-31'),
  (2, 'タスク2', 2, 'medium', '2025-11-30'),
  (3, 'タスク3', 3, 'low', NULL);


