BINARY=seedling
BUILD_DIR=build
GO=go
GOFLAGS=-ldflags="-s -w"

.PHONY: all build test test-verbose vet lint clean run help

all: vet test build

build:
	mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/$(BINARY)

test:
	$(GO) test -count=1 ./...

test-verbose:
	$(GO) test -v -count=1 ./...

vet:
	$(GO) vet ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BUILD_DIR)

run:
	$(GO) run ./cmd/$(BINARY)

help:
	@echo "Targets:"
	@echo "  all           - vet -> test -> build"
	@echo "  build         - compile binary into $(BUILD_DIR)/"
	@echo "  test          - run all tests"
	@echo "  test-verbose  - run all tests with verbose output"
	@echo "  vet           - run go vet"
	@echo "  lint          - run golangci-lint"
	@echo "  clean         - remove $(BUILD_DIR)/"
	@echo "  run           - go run the CLI"
	@echo ""
	@echo "Example:"
	@echo "  make build && ./$(BUILD_DIR)/$(BINARY) introspect --help"
