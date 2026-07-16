.PHONY: build test test-race lint install run tidy clean

BIN     := lexicon
PKG     := ./cmd/lexicon
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN) $(PKG)

test:
	go test -count=1 ./...

test-race:
	go test -race -count=1 ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint not installed — see https://golangci-lint.run/welcome/install/"; \
		exit 1; \
	}
	golangci-lint run ./...

install:
	go install -trimpath -ldflags "$(LDFLAGS)" $(PKG)

run: build
	./$(BIN) compile --list-targets

tidy:
	go mod tidy

clean:
	rm -rf $(BIN) $(BIN).exe dist coverage coverage.out
