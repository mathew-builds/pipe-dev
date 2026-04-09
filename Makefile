.PHONY: build run demo test lint clean

BINARY=pipe
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X github.com/mathew-builds/pipe-dev/pkg/version.Version=$(VERSION)"

build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/pipe

run: build
	./$(BINARY)

demo: build
	./$(BINARY) demo

test:
	go test ./... -v

lint:
	golangci-lint run

clean:
	rm -f $(BINARY)
