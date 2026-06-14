BINARY_NAME=tendr
VERSION=0.1.0
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(shell date +%FT%T%z)
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE)"

.PHONY: all build test lint clean release install

all: build

build:
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/tendr

test:
	CGO_ENABLED=0 go test ./... -v -count=1

lint:
	CGO_ENABLED=0 go vet ./...

clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

release:
	goreleaser release --snapshot --clean

install: build
	cp $(BINARY_NAME) $(shell go env GOPATH)/bin/$(BINARY_NAME)
