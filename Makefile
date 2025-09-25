# HttpGo Makefile
# 用于构建、测试和发布 HttpGo

# 项目信息
PROJECT_NAME := httpgo
PKG := github.com/itnxs/httpgo
VERSION_PKG := $(PKG)/internal/pkg

# 版本信息
VERSION ?= $(shell git describe --tags --exact-match HEAD 2>/dev/null || echo "v1.0.0-dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go 相关变量
GO := go
GOOS ?= $(shell $(GO) env GOOS)
GOARCH ?= $(shell $(GO) env GOARCH)
CGO_ENABLED ?= 0

# 构建标志
LDFLAGS := -s -w \
	-X '$(VERSION_PKG).Version=$(VERSION)' \
	-X '$(VERSION_PKG).Commit=$(COMMIT)' \
	-X '$(VERSION_PKG).BuildTime=$(BUILD_TIME)'

# 构建目录
BUILD_DIR := build
DIST_DIR := dist

# 支持的平台
PLATFORMS := \
	windows/amd64 \
	windows/386 \
	windows/arm64 \
	linux/amd64 \
	linux/386 \
	linux/arm64 \
	linux/arm \
	darwin/amd64 \
	darwin/arm64 \
	freebsd/amd64

# 默认目标
.PHONY: all
all: test build

# 显示帮助信息
.PHONY: help
help:
	@echo "HttpGo 构建系统"
	@echo ""
	@echo "可用目标:"
	@echo "  help            显示此帮助信息"
	@echo "  clean           清理构建文件"
	@echo "  deps            获取依赖"
	@echo "  test            运行测试"
	@echo "  build           构建当前平台版本"
	@echo "  build-all       构建所有平台版本"
	@echo "  dist            创建发布包"
	@echo "  release         创建 GitHub Release"
	@echo "  install         安装到本地"
	@echo "  run             运行程序"
	@echo ""
	@echo "变量:"
	@echo "  VERSION         版本号 (当前: $(VERSION))"
	@echo "  GOOS            目标操作系统"
	@echo "  GOARCH          目标架构"
	@echo ""
	@echo "示例:"
	@echo "  make build                    # 构建当前平台"
	@echo "  make build GOOS=linux         # 构建 Linux 版本"
	@echo "  make build-all                # 构建所有平台"
	@echo "  make VERSION=v1.2.0 build-all # 指定版本构建"

# 清理
.PHONY: clean
clean:
	@echo "清理构建文件..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	@echo "清理完成"

# 获取依赖
.PHONY: deps
deps:
	@echo "获取依赖..."
	@$(GO) mod download
	@$(GO) mod tidy
	@echo "依赖获取完成"

# 运行测试
.PHONY: test
test:
	@echo "运行测试..."
	@$(GO) test -v ./...
	@echo "测试完成"

# 运行基准测试
.PHONY: bench
bench:
	@echo "运行基准测试..."
	@$(GO) test -bench=. -benchmem ./...

# 代码检查
.PHONY: lint
lint:
	@echo "运行代码检查..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint 未安装，跳过代码检查"; \
	fi

# 构建当前平台
.PHONY: build
build: clean
	@echo "构建 $(GOOS)/$(GOARCH) 版本..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) \
		$(GO) build -ldflags="$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(PROJECT_NAME)$(shell [ "$(GOOS)" = "windows" ] && echo ".exe") .
	@echo "构建完成: $(BUILD_DIR)/$(PROJECT_NAME)$(shell [ "$(GOOS)" = "windows" ] && echo ".exe")"

# 构建所有平台
.PHONY: build-all
build-all: clean
	@echo "构建所有平台版本..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		suffix=""; \
		[ "$$os" = "windows" ] && suffix=".exe"; \
		echo "构建 $$os/$$arch..."; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch \
			$(GO) build -ldflags="$(LDFLAGS)" \
			-o $(BUILD_DIR)/$(PROJECT_NAME)-$$os-$$arch$$suffix .; \
		if [ $$? -eq 0 ]; then \
			echo "✓ $$os/$$arch 构建成功"; \
		else \
			echo "✗ $$os/$$arch 构建失败"; \
		fi; \
	done
	@echo "所有平台构建完成"

# 创建发布包
.PHONY: dist
dist: build-all
	@echo "创建发布包..."
	@mkdir -p $(DIST_DIR)
	@cd $(BUILD_DIR) && \
	for file in $(PROJECT_NAME)-*; do \
		if [ -f "$$file" ]; then \
			platform=$$(echo $$file | sed 's/$(PROJECT_NAME)-//'); \
			if echo "$$platform" | grep -q windows; then \
				zip "$(DIST_DIR)/$$file.zip" "$$file"; \
				echo "创建: $$file.zip"; \
			else \
				tar -czf "$(DIST_DIR)/$$file.tar.gz" "$$file"; \
				echo "创建: $$file.tar.gz"; \
			fi; \
		fi; \
	done
	@echo "生成校验和..."
	@cd $(DIST_DIR) && \
	if command -v sha256sum >/dev/null 2>&1; then \
		sha256sum * > SHA256SUMS; \
	elif command -v shasum >/dev/null 2>&1; then \
		shasum -a 256 * > SHA256SUMS; \
	fi
	@echo "发布包创建完成，文件位于 $(DIST_DIR)/ 目录"

# 安装到本地
.PHONY: install
install:
	@echo "安装到本地..."
	@$(GO) install -ldflags="$(LDFLAGS)" .
	@echo "安装完成"

# 运行程序
.PHONY: run
run:
	@$(GO) run -ldflags="$(LDFLAGS)" . $(ARGS)

# 开发模式运行
.PHONY: dev
dev:
	@$(GO) run . --help

# 创建 git tag
.PHONY: tag
tag:
	@if [ -z "$(TAG)" ]; then \
		echo "请指定 TAG，例如: make tag TAG=v1.0.0"; \
		exit 1; \
	fi
	@echo "创建标签: $(TAG)"
	@git tag -a $(TAG) -m "Release $(TAG)"
	@echo "推送标签到远程仓库..."
	@git push origin $(TAG)

# 删除 git tag
.PHONY: untag
untag:
	@if [ -z "$(TAG)" ]; then \
		echo "请指定 TAG，例如: make untag TAG=v1.0.0"; \
		exit 1; \
	fi
	@echo "删除本地标签: $(TAG)"
	@git tag -d $(TAG) || true
	@echo "删除远程标签: $(TAG)"
	@git push origin :refs/tags/$(TAG) || true

# 显示版本信息
.PHONY: version
version:
	@echo "项目: $(PROJECT_NAME)"
	@echo "版本: $(VERSION)"
	@echo "提交: $(COMMIT)"
	@echo "构建时间: $(BUILD_TIME)"
	@echo "Go 版本: $(shell $(GO) version)"

# 显示构建信息
.PHONY: info
info:
	@echo "构建信息:"
	@echo "  项目名称: $(PROJECT_NAME)"
	@echo "  包路径: $(PKG)"
	@echo "  版本: $(VERSION)"
	@echo "  提交: $(COMMIT)"
	@echo "  构建时间: $(BUILD_TIME)"
	@echo "  目标平台: $(GOOS)/$(GOARCH)"
	@echo "  CGO: $(CGO_ENABLED)"
	@echo "  构建标志: $(LDFLAGS)"

# 快速发布流程
.PHONY: release
release: test build-all dist
	@echo "发布准备完成！"
	@echo ""
	@echo "接下来的步骤:"
	@echo "1. 检查构建文件: ls -la $(DIST_DIR)/"
	@echo "2. 创建 git tag: make tag TAG=v1.x.x"
	@echo "3. 推送 tag 将自动触发 GitHub Actions 创建 Release"
	@echo ""
	@echo "或者手动上传 $(DIST_DIR)/ 中的文件到 GitHub Releases"

# 检查是否有未提交的更改
.PHONY: check-git
check-git:
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "警告: 有未提交的更改"; \
		git status --short; \
		exit 1; \
	fi
	@echo "Git 状态检查通过"