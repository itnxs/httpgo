# HttpGo Release 发布指南

本文档介绍如何为 HttpGo 项目创建和发布新版本。

## 🚀 快速发布

### 1. 自动发布（推荐）

使用 GitHub Actions 自动构建和发布：

```bash
# 1. 确保所有更改已提交
git add .
git commit -m "feat: 准备发布 v1.0.0"

# 2. 创建并推送标签
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# GitHub Actions 将自动：
# - 构建所有平台版本
# - 运行测试
# - 创建 GitHub Release
# - 上传构建文件
```

### 2. 手动发布

#### 使用 Makefile（Linux/macOS）

```bash
# 构建所有平台并创建发布包
make release

# 或分步执行
make test           # 运行测试
make build-all      # 构建所有平台
make dist          # 创建发布包

# 查看构建结果
ls -la dist/
```

#### 使用构建脚本

**Linux/macOS:**
```bash
# 给脚本执行权限
chmod +x scripts/build.sh

# 构建指定版本
./scripts/build.sh v1.0.0

# 或使用默认版本
./scripts/build.sh
```

**Windows:**
```cmd
# 运行 Windows 构建脚本
scripts\build.bat v1.0.0
```

## 📋 发布流程详解

### 准备阶段

1. **更新版本号**
   ```bash
   # 版本号在 internal/pkg/httpgo.go 中定义
   const Version = "1.0.0"
   ```

2. **更新 CHANGELOG**（如果有）
   ```markdown
   ## [1.0.0] - 2024-01-01
   ### Added
   - 新功能描述
   ### Fixed
   - 修复的问题
   ```

3. **确保测试通过**
   ```bash
   go test ./...
   ```

### 构建阶段

构建系统支持以下平台：

| 操作系统 | 架构 | 文件名示例 |
|----------|------|------------|
| Windows | amd64 | httpgo-windows-amd64.exe |
| Windows | 386 | httpgo-windows-386.exe |
| Windows | arm64 | httpgo-windows-arm64.exe |
| Linux | amd64 | httpgo-linux-amd64 |
| Linux | 386 | httpgo-linux-386 |
| Linux | arm64 | httpgo-linux-arm64 |
| Linux | arm | httpgo-linux-arm |
| macOS | amd64 | httpgo-darwin-amd64 |
| macOS | arm64 | httpgo-darwin-arm64 |
| FreeBSD | amd64 | httpgo-freebsd-amd64 |

### 发布阶段

#### GitHub Releases

1. **自动发布**（推荐）
   - 推送标签到 GitHub 自动触发
   - GitHub Actions 自动创建 Release

2. **手动发布**
   - 访问 GitHub Releases 页面
   - 点击 "Create a new release"
   - 上传构建的压缩包
   - 添加 Release 说明

## 🛠️ 构建选项

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `GOOS` | 目标操作系统 | 当前系统 |
| `GOARCH` | 目标架构 | 当前架构 |
| `CGO_ENABLED` | 是否启用 CGO | 0 |
| `VERSION` | 版本号 | 从 git tag 获取 |

### 构建标志

构建时会注入以下信息：
- 版本号
- Git 提交哈希
- 构建时间

```bash
# 查看版本信息
./httpgo --version
```

## 📦 发布包结构

```
dist/
├── httpgo-windows-amd64.zip
├── httpgo-windows-386.zip
├── httpgo-windows-arm64.zip
├── httpgo-linux-amd64.tar.gz
├── httpgo-linux-386.tar.gz
├── httpgo-linux-arm64.tar.gz
├── httpgo-linux-arm.tar.gz
├── httpgo-darwin-amd64.tar.gz
├── httpgo-darwin-arm64.tar.gz
├── httpgo-freebsd-amd64.tar.gz
└── SHA256SUMS                    # 校验和文件
```

## 🔧 自定义构建

### 单平台构建

```bash
# 构建 Linux 64位版本
GOOS=linux GOARCH=amd64 go build -o httpgo-linux-amd64

# 构建 Windows 版本
GOOS=windows GOARCH=amd64 go build -o httpgo-windows-amd64.exe
```

### 添加新平台

在以下文件中添加新的构建目标：

1. **GitHub Actions**: `.github/workflows/release.yml`
2. **构建脚本**: `scripts/build.sh`
3. **Makefile**: `Makefile`

示例添加 ARM 架构：
```yaml
# GitHub Actions
- goos: linux
  goarch: arm
  suffix: ''
  name: httpgo-linux-arm
```

## 🐛 故障排除

### 常见问题

1. **构建失败**
   ```bash
   # 检查 Go 环境
   go version
   go env
   
   # 清理并重试
   go clean -cache
   make clean
   make build-all
   ```

2. **依赖问题**
   ```bash
   # 更新依赖
   go mod tidy
   go mod download
   ```

3. **权限问题**
   ```bash
   # Linux/macOS 赋予执行权限
   chmod +x scripts/build.sh
   ```

### 调试构建

启用详细输出：
```bash
# Makefile 详细模式
make build-all VERBOSE=1

# 构建脚本详细模式
./scripts/build.sh --verbose
```

## 📈 发布最佳实践

1. **版本管理**
   - 使用语义化版本号（SemVer）
   - 格式：`v<major>.<minor>.<patch>`
   - 示例：`v1.0.0`, `v1.2.3`, `v2.0.0-beta.1`

2. **测试验证**
   - 发布前运行完整测试套件
   - 在不同平台上验证构建结果
   - 检查二进制文件是否正常运行

3. **文档更新**
   - 更新 README.md
   - 添加 CHANGELOG 条目
   - 更新版本相关文档

4. **备份策略**
   - 保留发布构建文件
   - 记录构建环境信息
   - 备份源代码快照

## 🔄 回滚发布

如果需要回滚发布：

```bash
# 删除错误的标签
git tag -d v1.0.0
git push origin :refs/tags/v1.0.0

# 在 GitHub 上删除 Release

# 重新创建正确的标签
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

## 📞 获取帮助

如果在发布过程中遇到问题：

1. 查看构建日志
2. 检查 GitHub Actions 输出
3. 在项目 Issues 中报告问题
4. 联系维护团队