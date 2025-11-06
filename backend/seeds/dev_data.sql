-- テストユーザー
INSERT INTO users (email, password_hash, name, role) VALUES
  ('admin@example.com', '$hashed_password', 'Admin User', 'admin'),
  ('user1@example.com', '$hashed_password', 'Test User 1', 'user'),
  ('user2@example.com', '$hashed_password', 'Test User 2', 'user');

-- テストタスク
INSERT INTO tasks (user_id, title, description, status, priority, due_date) VALUES
  (2, 'タスク1', '説明1', 'todo', 'high', '2025-12-31'),
  (2, 'タスク2', '説明2', 'in_progress', 'medium', '2025-11-30'),
  (3, 'タスク3', '説明3', 'done', 'low', NULL);
