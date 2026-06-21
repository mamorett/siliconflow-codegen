MAKEFLAGS += --no-print-directory
BINARY_NAME := siliconflow-codegen
DIST_DIR := dist
OPENCODE_CONFIG ?= siliconflow.opencode.json
CRUSH_CONFIG ?= siliconflow.crush.json
QWENCODE_CONFIG ?= siliconflow.qwencode.json
LDFLAGS ?= -s -w

LOCAL_BINARY := $(DIST_DIR)/$(BINARY_NAME)

BINARIES := \
	$(DIST_DIR)/$(BINARY_NAME)-linux-arm64 \
	$(DIST_DIR)/$(BINARY_NAME)-linux-amd64 \
	$(DIST_DIR)/$(BINARY_NAME)-darwin-arm64

.PHONY: all build dist clean gen-opencode gen-crush gen-qwencode claude gen-opencode-linux-arm64 gen-opencode-linux-amd64 gen-opencode-darwin-arm64 models raw-models test format help

all: build

build:
	@mkdir -p $(DIST_DIR)
	go build -trimpath -ldflags="$(LDFLAGS)" -o $(LOCAL_BINARY) .

dist: $(BINARIES)

$(DIST_DIR)/$(BINARY_NAME)-linux-arm64:
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" -o $@ .

$(DIST_DIR)/$(BINARY_NAME)-linux-amd64:
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" -o $@ .

$(DIST_DIR)/$(BINARY_NAME)-darwin-arm64:
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" -o $@ .

gen-opencode:
	@test -n "$${SILICONFLOW_API_KEY}" || { echo "ERROR: SILICONFLOW_API_KEY is required" >&2; exit 1; }
	go run . --gen-opencode > $(OPENCODE_CONFIG)

gen-crush:
	@test -n "$${SILICONFLOW_API_KEY}" || { echo "ERROR: SILICONFLOW_API_KEY is required" >&2; exit 1; }
	go run . --gen-crush > $(CRUSH_CONFIG)

gen-qwencode:
	@test -n "$${SILICONFLOW_API_KEY}" || { echo "ERROR: SILICONFLOW_API_KEY is required" >&2; exit 1; }
	go run . --gen-qwencode > $(QWENCODE_CONFIG)

claude:
	@test -n "$${SILICONFLOW_API_KEY}" || { echo "ERROR: SILICONFLOW_API_KEY is required" >&2; exit 1; }
	@go run . --claude

gen-opencode-linux-arm64: $(DIST_DIR)/$(BINARY_NAME)-linux-arm64
	@test -n "$${SILICONFLOW_API_KEY}" || { echo "ERROR: SILICONFLOW_API_KEY is required" >&2; exit 1; }
	./$(DIST_DIR)/$(BINARY_NAME)-linux-arm64 --gen-opencode > $(OPENCODE_CONFIG)

gen-opencode-linux-amd64: $(DIST_DIR)/$(BINARY_NAME)-linux-amd64
	@test -n "$${SILICONFLOW_API_KEY}" || { echo "ERROR: SILICONFLOW_API_KEY is required" >&2; exit 1; }
	./$(DIST_DIR)/$(BINARY_NAME)-linux-amd64 --gen-opencode > $(OPENCODE_CONFIG)

gen-opencode-darwin-arm64: $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64
	@test -n "$${SILICONFLOW_API_KEY}" || { echo "ERROR: SILICONFLOW_API_KEY is required" >&2; exit 1; }
	./$(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 --gen-opencode > $(OPENCODE_CONFIG)

raw-models:
	@test -n "$${SILICONFLOW_API_KEY}" || { echo "ERROR: SILICONFLOW_API_KEY is required" >&2; exit 1; }
	go run . > siliconflow.models.json

models: raw-models

test:
	go test ./...

format:
	gofmt -w siliconflow-codegen.go

clean:
	rm -rf $(DIST_DIR) siliconflow.opencode.json siliconflow.crush.json siliconflow.qwencode.json siliconflow.models.json

help:
	@echo "siliconflow-codegen"
	@echo ""
	@echo "Targets:"
	@echo "  build                         Build the native binary"
	@echo "  dist                          Build linux-arm64, linux-amd64 and darwin-arm64 binaries"
	@echo "  gen-opencode                  Generate $(OPENCODE_CONFIG) with the local Go toolchain"
	@echo "  gen-crush                     Generate $(CRUSH_CONFIG) with the local Go toolchain"
	@echo "  gen-qwencode                  Generate $(QWENCODE_CONFIG) with the local Go toolchain"
	@echo "  claude                        Interactively select a SiliconFlow model, update Claude Code settings, and print exports"
	@echo "  gen-opencode-linux-arm64      Build linux-arm64 binary and generate $(OPENCODE_CONFIG)"
	@echo "  gen-opencode-linux-amd64      Build linux-amd64 binary and generate $(OPENCODE_CONFIG)"
	@echo "  gen-opencode-darwin-arm64     Build darwin-arm64 binary and generate $(OPENCODE_CONFIG)"
	@echo "  raw-models                    Fetch and save the raw SiliconFlow model API response"
	@echo "  test                          Run Go tests"
	@echo "  format                        Format Go source"
	@echo "  clean                         Remove generated binaries and config files"