# Makefile for Git Log Analyzer

# 变量定义
BINARY_NAME=git-log-analyzer
MAIN_PATH=./main.go
BUILD_DIR=./build
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# 默认目标
.PHONY: all
all: clean deps build

# 安装依赖
.PHONY: deps
deps:
	go mod download
	go mod tidy

# 构建
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# 构建多平台版本
.PHONY: build-all
build-all: clean deps
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	
	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	
	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	
	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	
	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

# 安装到系统
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	sudo chmod +x /usr/local/bin/$(BINARY_NAME)

# 运行测试
.PHONY: test
test:
	go test -v ./...

# 运行测试并生成覆盖率报告
.PHONY: test-coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 格式化代码
.PHONY: fmt
fmt:
	go fmt ./...

# 静态检查
.PHONY: vet
vet:
	go vet ./...

# 运行linter
.PHONY: lint
lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

# 清理构建产物
.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# 开发环境运行
.PHONY: dev
dev:
	go run $(MAIN_PATH) --help

# 运行示例分析
.PHONY: example
example: build
	@echo "Running example analysis on current repository..."
	$(BUILD_DIR)/$(BINARY_NAME) --repo .

# 检查代码质量
.PHONY: check
check: fmt vet lint test

# 发布准备
.PHONY: release
release: clean check build-all
	@echo "Release artifacts built in $(BUILD_DIR)/"
	@ls -la $(BUILD_DIR)/

# 显示帮助
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all          - Clean, install deps, and build"
	@echo "  deps         - Install dependencies"
	@echo "  build        - Build binary for current platform"
	@echo "  build-all    - Build binaries for all platforms"
	@echo "  install      - Install binary to system"
	@echo "  test         - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  fmt          - Format code"
	@echo "  vet          - Run go vet"
	@echo "  lint         - Run golangci-lint"
	@echo "  clean        - Clean build artifacts"
	@echo "  dev          - Run in development mode"
	@echo "  example      - Run example analysis"
	@echo "  check        - Run all code quality checks"
	@echo "  release      - Prepare release artifacts"
	@echo "  help         - Show this help"
# Build optimization flags
go build -ldflags "-w -s" -o git-log-analyzer
