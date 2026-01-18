CREATE TABLE IF NOT EXISTS user_roles (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(30) NOT NULL CHECK(name IN('admin','writer')) ,
  display_order INTEGER NOT NULL,     -- UI表示順序 (例: 1, 2, 3)
  is_active BOOLEAN NOT NULL DEFAULT true,  -- 論理削除フラグ
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP DEFAULT NULL
);

CREATE UNIQUE INDEX idx_user_roles_name ON user_roles(name);
CREATE INDEX idx_user_roles_display_order ON user_roles(display_order);

-- 初期データ
INSERT INTO user_roles (name, display_order) VALUES
  ('admin', 1),
  ('writer', 2)
ON CONFLICT DO NOTHING;
