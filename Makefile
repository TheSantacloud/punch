format:
	go fmt ./...

build:
	go build -o ~/.local/bin/flt

.DEFAULT_GOAL := all
all: format build 
