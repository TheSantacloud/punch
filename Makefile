format:
	go fmt ./...

build:
	go build -o ~/.local/bin/punch

test:
	go test ./...

.PHONY: all

.DEFAULT_GOAL := all
all: format test build
