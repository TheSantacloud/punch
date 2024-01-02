format:
	go fmt ./...

build:
	go build -o ~/.local/bin/punch

test:
	go test ./...

lint:
	golangci-lint run

.PHONY: all

.DEFAULT_GOAL := all
all: format lint test build
