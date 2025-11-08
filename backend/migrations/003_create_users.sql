BEGIN

DROP TABLE IF EXISTS users;

CREATE TABLE IF NOT EXISTS users (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  email VARCHAR(100) NOT NULL,
  password_hash VARCHAR(500) NOT NULL,
  role_id BIGINT NOT NULL DEFAULT 1,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP DEFAULT NULL,

  CONSTRAINT fk_users_role_id 
  FOREIGN KEY (role_id) 
  REFERENCES user_roles(id) 
  ON DELETE RESTRICT
);

CREATE INDEX idx_users_name ON users(name);
CREATE UNIQUE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role_id ON users(role_id);

COMMIT;

-- 変更を取り消したい場合に使用する。
-- ROLLBACK;
