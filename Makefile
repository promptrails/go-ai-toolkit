.PHONY: build run clean install fmt vet lint

## Build the binary
build:
	go build -o ai-chat ./cmd/chat/

## Run the app
run: build
	./ai-chat

## Format code
fmt:
	gofmt -w .

## Run go vet
vet:
	go vet ./...

## Run linters
lint:
	golangci-lint run ./...

## Clean build artifacts
clean:
	rm -f ai-chat

## Install to GOPATH/bin
install:
	go install ./cmd/chat/
