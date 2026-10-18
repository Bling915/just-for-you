GO ?= go
MYSQL_DSN ?= just_for_you:local-development@tcp(127.0.0.1:3306)/just_for_you?parseTime=true&multiStatements=true

.PHONY: test build-server migrate-up migrate-down
test:
	cd server && $(GO) test ./...
build-server:
	cd server && $(GO) build ./cmd/api
migrate-up:
	cd server && $(GO) run ./cmd/migrate -direction up
migrate-down:
	cd server && $(GO) run ./cmd/migrate -direction down
