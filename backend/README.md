
```zsh
# 1. Go モジュール初期化
go mod init github.com/take/gin-todo

# 2. 基本パッケージ
go get -u github.com/gin-gonic/gin \
  gorm.io/gorm \
  gorm.io/driver/postgres \
  github.com/golang-jwt/jwt/v5 \
  github.com/joho/godotenv \
  github.com/go-playground/validator/v10 \
  github.com/gorilla/websocket \
  github.com/swaggo/swag \
  github.com/swaggo/gin-swagger \
  github.com/swaggo/files

# 3. Swagger
go get -u github.com/swaggo/swag \
  github.com/swaggo/gin-swagger \
  github.com/swaggo/files

# 4. swagコマンドインストール
go install github.com/swaggo/swag/cmd/swag@latest

# 5. 依存関係整理
go mod tidy
```
