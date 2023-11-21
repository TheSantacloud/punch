format:
	go fmt ./...

build:
	go build -o ~/.local/bin/punch

.DEFAULT_GOAL := all
all: format build 
