.DEFAULT_GOAL := build

fmt:
	go fmt ./...
.PHONY: fmt

lint: fmt
	golint ./...
.PHONY: lint

vet: fmt
	go vet ./...
.PHONY: vet

build:
	go build hello.go
.PHONY: build

test:
	go test -v ./...
.PHONY: test

run-serve:
	CGO_ENABLED=1 go build cmd/serve/serve.go
	CGO_ENABLED=1 go run cmd/serve/serve.go
.PHONY: run-serve

run-ingest:
	CGO_ENABLED=1 go run cmd/ingest/ingest.go
.PHONY: run-ingest

