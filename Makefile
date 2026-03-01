TOOLS_DIR := .tools
GOBIN := $(abspath $(TOOLS_DIR))
export GOBIN
CACHE_DIR := .cache
GOCACHE := $(abspath $(CACHE_DIR)/go-build)
GOMODCACHE := $(abspath $(CACHE_DIR)/go-mod)
export GOCACHE
export GOMODCACHE
GO := GO111MODULE=on GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) go

BINARY := papercli
CMD_DIR := ./cmd/papercli
BIN_DIR := bin
OUTPUT := $(BIN_DIR)/$(BINARY)
VERSION ?=
LDFLAGS :=
ifneq ($(strip $(VERSION)),)
LDFLAGS := -ldflags "-X main.version=$(VERSION)"
endif

.DEFAULT_GOAL := build

GOFILES := $(shell find . -name '*.go' -not -path './.tools/*' -not -path './.cache/*' -not -path './vendor/*')

.PHONY: build clean tools fmt fmt-check lint test run

build:
	@mkdir -p $(BIN_DIR) $(GOCACHE) $(GOMODCACHE)
	@$(GO) build $(LDFLAGS) -o $(OUTPUT) $(CMD_DIR)
	@echo "built $(OUTPUT)"

run: build
	@$(OUTPUT) --version

clean:
	@rm -rf $(BIN_DIR) $(CACHE_DIR)

tools:
	@mkdir -p $(TOOLS_DIR) $(GOCACHE) $(GOMODCACHE)
	@$(GO) install mvdan.cc/gofumpt@v0.7.0
	@$(GO) install golang.org/x/tools/cmd/goimports@v0.38.0
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.62.2

fmt:
	@$(GOBIN)/goimports -w $(GOFILES)
	@$(GOBIN)/gofumpt -w $(GOFILES)

fmt-check:
	@$(GOBIN)/goimports -w $(GOFILES)
	@$(GOBIN)/gofumpt -w $(GOFILES)
	@git diff --exit-code -- '*.go' go.mod go.sum

lint:
	@$(GOBIN)/golangci-lint run ./...

test:
	@$(GO) test ./...
