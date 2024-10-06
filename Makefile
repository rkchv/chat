include .env
LOCAL_BIN:=$(CURDIR)/bin
LOCAL_MIGRATION_DSN="host=localhost port=$(PG_PORT) dbname=$(PG_DBNAME) user=$(PG_USER) password=$(PG_PWD) sslmode=disable"

lint:
	GOBIN=$(LOCAL_BIN) golangci-lint run ./... --config .golangci.pipeline.yaml

install-deps:
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.53.3
	GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
	GOBIN=$(LOCAL_BIN) go install -mod=mod google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.4
	GOBIN=$(LOCAL_BIN) go install -mod=mod github.com/pressly/goose/v3/cmd/goose@v3.21.1

get-deps:
	go get -u google.golang.org/protobuf/cmd/protoc-gen-go
	go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc
	go get -u github.com/jackc/pgx/v5
	go get -u golang.org/x/exp/slog
	go get -u github.com/Masterminds/squirrel

generate:
	mkdir -p pkg/chat_v1
	protoc --proto_path api/chat_v1 \
	--go_out=pkg/chat_v1 --go_opt=paths=source_relative \
	--plugin=protoc-gen-go=bin/protoc-gen-go \
	--go-grpc_out=pkg/chat_v1 --go-grpc_opt=paths=source_relative \
	--plugin=protoc-gen-go-grpc=bin/protoc-gen-go-grpc \
	api/chat_v1/chat.proto

build:
	GOOS=linux GOARCH=amd64 go build -o service_linux cmd/grpc_server/main.go

local-up:
	$(LOCAL_BIN)/goose -dir $(MIGRATION_DIR) postgres ${LOCAL_MIGRATION_DSN} up -v
local-down:
	$(LOCAL_BIN)/goose -dir $(MIGRATION_DIR) postgres ${LOCAL_MIGRATION_DSN} down -v

run:
	docker compose up -d
