.PHONY: swagger
swagger:
	swag init -g cmd/api/main.go

.PHONY: run
run:
	go run cmd/api/main.go

.PHONY: dev
dev: swagger run
