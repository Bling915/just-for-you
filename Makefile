LOCAL_GO := $(CURDIR)/.superpowers/tools/go/bin/go
GO_BIN := $(shell if [ -x "$(LOCAL_GO)" ]; then printf '%s' "$(LOCAL_GO)"; else command -v go; fi)
MYSQL_DSN ?= just_for_you:local-development@tcp(127.0.0.1:3306)/just_for_you?parseTime=true&multiStatements=true

.PHONY: test verify test-backend test-miniprogram test-race vet build-server migrate-up migrate-down
test: test-backend test-miniprogram
verify: test-miniprogram test-backend test-race vet
test-backend:
	cd server && "$(GO_BIN)" test ./...
test-miniprogram:
	node --test miniprogram/tests/*.test.js
test-race:
	cd server && "$(GO_BIN)" test -race ./internal/identity/... ./internal/relationship/... ./internal/moment/...
vet:
	cd server && "$(GO_BIN)" vet ./...
build-server:
	cd server && "$(GO_BIN)" build ./cmd/api
migrate-up:
	cd server && "$(GO_BIN)" run ./cmd/migrate -direction up
migrate-down:
	cd server && "$(GO_BIN)" run ./cmd/migrate -direction down
