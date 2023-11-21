format:
	go fmt ./...

build:
	go build -o ~/.local/bin/work

.DEFAULT_GOAL := all
all: format build 
