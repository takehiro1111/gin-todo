
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

-- パスワードは全て 'password123' のbcryptハッシュ
INSERT INTO users (email, password_hash, name, role_id) VALUES
  ('admin@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZRGdjGj/n3.0B1Qc6T5PQxqUzuSO2', 'Admin User', 1),
  ('user1@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZRGdjGj/n3.0B1Qc6T5PQxqUzuSO2', 'Test User 1', 2)

INSERT INTO tasks (user_id, title, status_id, priority, due_date) VALUES
  (14, 'タスク1', 1, 'high', NULL),
  (14, 'タスク2', 2, 'medium', NULL),
  (14, 'タスク3', 3, 'low', NULL);


