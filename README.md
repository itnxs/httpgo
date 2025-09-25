# HttpGo

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.18-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

HttpGo 大部分由AI编写完成, 由人类删除多余代码
HttpGo 是一个快速、轻量级的 HTTP 基准测试工具，用 Go 语言编写。它提供了丰富的功能来测试 HTTP 服务的性能，支持各种协议、代理和认证方式。

## ✨ 特性

- 🚀 **高性能**: 基于 fasthttp 构建，支持高并发测试
- 🔧 **灵活配置**: 支持自定义连接数、请求数、超时时间等参数
- 📊 **实时统计**: 提供实时的 TUI 界面显示测试进度和统计信息
- 🌐 **多协议支持**: 支持 HTTP/1.1 和 HTTPS
- 🔒 **安全选项**: 支持 TLS 客户端证书和不安全连接
- 🌍 **代理支持**: 支持 HTTP 和 SOCKS 代理
- 📝 **多种请求格式**: 支持 JSON、表单数据、自定义请求体
- 🔄 **重定向处理**: 支持自动跟随 30x 重定向
- 📈 **限流控制**: 支持 QPS 限制

## 📦 安装

### 从源码编译

确保你已经安装了 Go 1.18 或更高版本：

```bash
git clone https://github.com/itnxs/httpgo.git
cd httpgo
go build
```

### 直接运行

```bash
go run main.go [参数]
```

## 🚀 快速开始

### 基本用法

```bash
# 简单的 GET 请求测试
./httpgo https://www.example.com

# 指定并发连接数和请求数
./httpgo https://www.example.com -c 10 -n 100

# 指定测试持续时间
./httpgo https://www.example.com -c 10 -d 30s
```

### 简化地址格式

HttpGo 支持简化的地址格式：

```bash
# 端口号（自动添加 localhost）
./httpgo :8080

# 路径（自动添加 localhost）
./httpgo /api/users

# 等价于
./httpgo http://localhost:8080
./httpgo http://localhost/api/users
```

## 📚 详细用法

### 命令行参数

#### 基本参数

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--connections` | `-c` | 128 | 最大并发连接数 |
| `--requests` | `-n` | 0 | 请求总数（如果指定，则忽略 --duration） |
| `--duration` | `-d` | 10s | 测试持续时间 |
| `--timeout` | `-t` | 3s | 套接字/请求超时时间 |
| `--qps` | | 0 | 固定基准测试的最大 QPS 值 |

#### HTTP 参数

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--method` | `-X` | GET | HTTP 请求方法 |
| `--header` | `-H` | | HTTP 请求头（可重复使用） |
| `--host` | | | 覆盖请求主机名 |
| `--body` | `-b` | | HTTP 请求体字符串 |
| `--file` | `-f` | | 从文件路径读取 HTTP 请求体 |

#### 内容类型

| 参数 | 简写 | 说明 |
|------|------|------|
| `--json` | `-J` | 发送 JSON 请求 |
| `--form` | `-F` | 发送表单请求 |
| `--stream` | `-s` | 使用流式请求体以减少内存使用 |

#### 连接选项

| 参数 | 简写 | 说明 |
|------|------|------|
| `--disableKeepAlives` | `-a` | 禁用 HTTP keep-alive |
| `--pipeline` | `-p` | 使用 fasthttp 管道客户端 |
| `--insecure` | `-k` | 跳过 TLS 证书验证 |

#### 认证和代理

| 参数 | 说明 |
|------|------|
| `--cert` | 客户端 TLS 证书路径 |
| `--key` | 客户端 TLS 证书私钥路径 |
| `--httpProxy` | HTTP 代理地址 |
| `--socksProxy` | SOCKS 代理地址 |

#### 调试选项

| 参数 | 简写 | 说明 |
|------|------|------|
| `--debug` | `-D` | 只发送一次请求并显示详情 |
| `--follow` | | 在调试模式下跟随 30x 重定向 |
| `--maxRedirects` | | 最大重定向次数（默认 30） |

### 使用示例

#### 1. 基本性能测试

```bash
# 使用 128 个并发连接测试 10 秒
./httpgo https://api.example.com -c 128 -d 10s

# 发送 1000 个请求，使用 50 个并发连接
./httpgo https://api.example.com -c 50 -n 1000
```

#### 2. QPS 限制测试

```bash
# 限制 QPS 为 100
./httpgo https://api.example.com --qps 100 -d 30s
```

#### 3. 自定义请求头

```bash
# 添加单个请求头
./httpgo https://api.example.com -H "Authorization: Bearer token123"

# 添加多个请求头
./httpgo https://api.example.com \
  -H "Authorization: Bearer token123" \
  -H "Content-Type: application/json" \
  -H "User-Agent: MyApp/1.0"
```

#### 4. POST 请求

```bash
# 发送 JSON 数据
./httpgo https://api.example.com -X POST -J -b '{"name":"test","value":123}'

# 发送表单数据
./httpgo https://api.example.com -X POST -F -b "name=test&value=123"

# 从文件读取请求体
./httpgo https://api.example.com -X POST -f request.json
```

#### 5. 简化参数格式

HttpGo 支持简化的键值对参数：

```bash
# JSON 格式（使用 :=）
./httpgo :3000 -c1 -n5 name:=test age:=25 active:=true

# 等价于
./httpgo http://localhost:3000 -c1 -n5 \
  -H "Content-Type: application/json" \
  -b '{"name":"test","age":25,"active":true}'

# 表单格式（使用 =）
./httpgo :3000 -c1 -n5 name=test age=25

# 等价于
./httpgo http://localhost:3000 -c1 -n5 \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -b "name=test&age=25"
```

#### 6. HTTPS 和证书

```bash
# 跳过证书验证
./httpgo https://self-signed.example.com -k

# 使用客户端证书
./httpgo https://api.example.com \
  --cert client.crt \
  --key client.key
```

#### 7. 代理支持

```bash
# HTTP 代理
./httpgo https://api.example.com --httpProxy http://proxy.example.com:8080

# SOCKS 代理
./httpgo https://api.example.com --socksProxy socks5://proxy.example.com:1080

# 带认证的代理
./httpgo https://api.example.com --httpProxy http://user:pass@proxy.example.com:8080
```

#### 8. 调试模式

```bash
# 调试模式：只发送一次请求并显示详细信息
./httpgo https://api.example.com -D

# 跟随重定向
./httpgo https://api.example.com -D --follow --maxRedirects 5
```

#### 9. 性能优化选项

```bash
# 使用管道连接
./httpgo https://api.example.com -p

# 使用流式请求体（减少内存使用）
./httpgo https://api.example.com -s -f large-file.json

# 禁用 keep-alive
./httpgo https://api.example.com -a
```

## 📊 输出说明

### 实时统计界面

HttpGo 提供了美观的实时 TUI 界面，显示以下信息：

```
Benchmarking https://api.example.com with 128 connections
████████████████████████████████████████████████████████████████████ 100%

Requests:  1000/1000  Elapsed: 10.5s  Throughput: 2.31 MB

                    Avg        Stdev       Max
Reqs/sec           95.24       12.34      120.56
Latency           13.45ms      5.23ms     45.67ms

HTTP codes:
  1xx - 0, 2xx - 980, 3xx - 15, 4xx - 5, 5xx - 0

Errors:
  No errors
```

### 调试模式输出

使用 `-D` 参数时，会显示完整的请求和响应详情：

```bash
Connected to api.example.com(192.168.1.100:443)

GET /api/users HTTP/1.1
Host: api.example.com
User-Agent: httpgo/1.0.0
Authorization: Bearer token123

HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 156

{"users": [...]}
```

## 🔧 配置选项详解

### 并发控制

- `--connections`: 控制同时建立的 TCP 连接数
- `--requests`: 总请求数量，优先级高于 duration
- `--duration`: 测试持续时间
- `--qps`: QPS 限制，优先级最高

### 超时设置

- `--timeout`: 适用于连接建立和请求响应的超时时间

### 请求体处理

- `--body`: 直接在命令行指定请求体
- `--file`: 从文件读取请求体，支持大文件
- `--stream`: 流式发送，适合大请求体，减少内存占用

## ⚠️ 注意事项

1. **HTTP/2 支持**: 当前版本由于依赖兼容性问题，HTTP/2 功能暂时禁用
2. **并发限制**: 高并发测试时请注意系统的文件描述符限制
3. **网络影响**: 测试结果会受到网络延迟和带宽的影响
4. **服务器负载**: 请确保不会对目标服务器造成过大压力

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

## 🙏 致谢

- [fasthttp](https://github.com/valyala/fasthttp) - 高性能 HTTP 客户端
- [cobra](https://github.com/spf13/cobra) - 命令行界面框架
- [bubbletea](https://github.com/charmbracelet/bubbletea) - TUI 框架