format:
	go fmt ./...

build:
	go build -o ~/.local/bin/punch

test:
	go test ./...

lint:
	golangci-lint run

update:
	go get -u ./...
	go mod tidy

.PHONY: all

.DEFAULT_GOAL := all
all: format lint test build
