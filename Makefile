BINARY   := broom
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS  := -ldflags "-X main.version=$(VERSION)"
GOROOT   ?= /opt/homebrew/Cellar/go/1.23.4/libexec
GO       := GOROOT=$(GOROOT) go

.PHONY: build run install clean test vet lint release

## build: compile the binary into ./broom
build:
	$(GO) build $(LDFLAGS) -o $(BINARY) .

## run: build and run broom
run: build
	./$(BINARY)

## install: install broom to /usr/local/bin
install: build
	install -m 0755 $(BINARY) /usr/local/bin/$(BINARY)

## test: run all tests
test:
	$(GO) test ./...

## vet: run go vet
vet:
	$(GO) vet ./...

## lint: run staticcheck (install with: go install honnef.co/go/tools/cmd/staticcheck@latest)
lint:
	staticcheck ./...

## clean: remove build artifacts
clean:
	rm -f $(BINARY)

## release: cross-compile for macOS (arm64 + amd64) and Linux (amd64)
release:
	GOROOT=$(GOROOT) GOOS=darwin  GOARCH=arm64  go build $(LDFLAGS) -o dist/$(BINARY)-darwin-arm64 .
	GOROOT=$(GOROOT) GOOS=darwin  GOARCH=amd64  go build $(LDFLAGS) -o dist/$(BINARY)-darwin-amd64 .
	GOROOT=$(GOROOT) GOOS=linux   GOARCH=amd64  go build $(LDFLAGS) -o dist/$(BINARY)-linux-amd64  .

## help: list available targets
help:
	@grep -E '^## ' Makefile | sed 's/## //'
