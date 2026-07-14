.PHONY: build test lint clean install

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o build/netbridge ./cmd/

test:
	go test -race -cover ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf build/

install: build
	install -m 755 build/netbridge /usr/local/bin/netbridge

dev:
	go run ./cmd/

fmt:
	gofmt -s -w .

vet:
	go vet ./...

tidy:
	go mod tidy

all: fmt vet test build
