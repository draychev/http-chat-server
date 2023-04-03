#!make

.PHONY: build
build:
	CGO_ENABLED=0 go build -v -o ./bin/http-chat-server ./http-chat-server.go

.PHONY: run
run: build
	./bin/http-chat-server
