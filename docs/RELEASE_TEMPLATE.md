# 发布说明模板

## HttpGo v{VERSION} 发布

### 🎉 新功能
- 新增功能 A
- 新增功能 B

### 🐛 修复
- 修复问题 A
- 修复问题 B

### 🔧 改进
- 改进性能
- 优化体验

### 📦 下载

请根据您的操作系统选择对应的版本：

#### Windows
- [httpgo-windows-amd64.zip](链接) - Windows 64位 (推荐)
- [httpgo-windows-386.zip](链接) - Windows 32位
- [httpgo-windows-arm64.zip](链接) - Windows ARM64

#### Linux
- [httpgo-linux-amd64.tar.gz](链接) - Linux 64位 (推荐)
- [httpgo-linux-386.tar.gz](链接) - Linux 32位
- [httpgo-linux-arm64.tar.gz](链接) - Linux ARM64
- [httpgo-linux-arm.tar.gz](链接) - Linux ARM

#### macOS
- [httpgo-darwin-amd64.tar.gz](链接) - Intel Mac
- [httpgo-darwin-arm64.tar.gz](链接) - Apple Silicon Mac (M1/M2)

#### FreeBSD
- [httpgo-freebsd-amd64.tar.gz](链接) - FreeBSD 64位

### 🔧 安装与使用

1. 下载对应平台的压缩包
2. 解压缩到您希望的目录
3. 运行 `httpgo --help` 查看使用说明
4. 开始基准测试：`httpgo https://example.com -c 10 -n 100`

### 📋 校验和

下载后请使用 SHA256SUMS 文件验证文件完整性：

```bash
# Linux/macOS
sha256sum -c SHA256SUMS

# Windows
certutil -hashfile httpgo-windows-amd64.zip SHA256
```

### 🚀 快速开始

```bash
# 基本使用
httpgo https://api.example.com

# 指定并发和请求数
httpgo https://api.example.com -c 50 -n 1000

# JSON 请求
httpgo https://api.example.com -X POST -J -b '{"key":"value"}'

# 更多选项
httpgo --help
```

### ⚠️ 注意事项

- HTTP/2 功能暂时禁用（兼容性问题）
- 高并发测试时请注意系统资源限制
- 请确保不会对目标服务器造成过大压力

### 🐛 问题反馈

如有问题请在 [GitHub Issues](https://github.com/itnxs/httpgo/issues) 中反馈。

### 💬 社区

- GitHub: https://github.com/itnxs/httpgo
- Issues: https://github.com/itnxs/httpgo/issues

---

**完整更改日志**: [v{PREV_VERSION}...v{VERSION}](https://github.com/itnxs/httpgo/compare/v{PREV_VERSION}...v{VERSION})