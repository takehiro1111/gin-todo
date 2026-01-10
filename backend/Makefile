.PHONY: swagger
swagger:
	swag init -g cmd/api/main.go

.PHONY: run
run:
	go run cmd/api/main.go

.PHONY: dev
dev: swagger run

# 開発段階でのサーバ起動
run-dev:
	go run cmd/api/main.go

# 製品段階でのサーバ起動
run-release:
	GIN_MODE=release go run cmd/api/main.go
