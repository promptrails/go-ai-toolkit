.PHONY: build run clean

## Build the binary
build:
	go build -o ai-chat ./cmd/chat/

## Run the app
run: build
	./ai-chat

## Clean build artifacts
clean:
	rm -f ai-chat

## Install to GOPATH/bin
install:
	go install ./cmd/chat/
