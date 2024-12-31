all: run
.DEFAULT_GOAL := build

fmt:
	go fmt ./...
.PHONY: fmt

vet: fmt
	go vet ./...
.PHONY: vet

build:
	go build cmd/serve/serve.go
	go build cmd/ingest/ingest.go
.PHONY: build

test:
	go test ./...
.PHONY: test

run-serve:
	CGO_ENABLED=1 go run cmd/serve/serve.go &
.PHONY: run-serve

run-ingest:
	CGO_ENABLED=1 go run cmd/ingest/ingest.go &
.PHONY: run-ingest

kill-serve:
	pkill -f "CGO_ENABLED=1 go run cmd/serve/serve.go" || true

kill-ingest:
	pkill -f "CGO_ENABLED=1 go run cmd/ingest/ingest.go" || true

run: run-ingest run-serve
.PHONY: run




