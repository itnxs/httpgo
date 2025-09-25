package pkg

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

// HTTP 相关常量
const (
	mimeApplicationJSON = "application/json"
	mimeApplicationForm = "application/x-www-form-urlencoded"
)

// Config 保存 httpgo 的配置设置
type Config struct {
	// Connections 表示并发使用的 TCP 连接数
	Connections int
	// Count 是一轮基准测试中的请求数量
	Count int
	// Qps 指定固定基准测试的最高值，但实际 qps
	// 可能会低于此值
	Qps int
	// Duration 表示基准测试持续时间，如果指定了 Count 则忽略此项
	Duration time.Duration
	// Timeout 表示套接字/请求超时时间
	Timeout time.Duration
	// Url 是基准测试的目标地址
	Url string
	// Method 是 HTTP 请求方法
	Method string
	// Args 可以方便地为表单和 JSON 请求设置数据
	Args []string
	// Headers 表示 HTTP 请求头
	Headers []string
	// Host 可以覆盖请求中的 Host 头
	Host string
	// DisableKeepAlives 将 Connection 头设置为 'close'
	DisableKeepAlives bool
	// Body 是请求体
	Body string
	// File 表示从文件中读取请求体
	File string
	// Stream 表示使用流式请求体
	Stream bool
	// JSON 表示发送 JSON 请求
	JSON bool
	// Form 表示发送表单请求
	Form bool
	// Insecure 跳过 TLS 验证
	Insecure bool
	// Cert 表示客户端 TLS 证书的路径
	Cert string
	// Key 表示客户端 TLS 证书私钥的路径
	Key string
	// HttpProxy 表示 HTTP 代理地址
	HttpProxy string
	// SocksProxy 表示 SOCKS 代理地址
	SocksProxy string
	// Pipeline 如果为 true，将使用 fasthttp PipelineClient
	Pipeline bool
	// Follow 如果为 true，在调试模式下跟随 30x 位置重定向
	Follow bool
	// MaxRedirects 表示跟随 30x 重定向的最大次数，
	// 默认为 30（仅在 Follow 为 true 时生效）
	MaxRedirects int
	// Debug 如果为 true，只发送一次请求并显示请求和响应详情
	Debug bool
	// Http2 如果为 true，将为 fasthttp 使用 http2
	Http2 bool

	throughput int64
	body       []byte
	isTLS      bool
	addr       string
	tlsConf    *tls.Config
}

func (c *Config) doer() (clientDoer, error) {
	if c.Pipeline {
		return &fasthttp.PipelineClient{
			Name:        "httpgo/" + Version,
			Addr:        c.addr,
			Dial:        c.getDialer(),
			IsTLS:       c.isTLS,
			TLSConfig:   c.tlsConf,
			MaxConns:    c.Connections,
			ReadTimeout: c.Timeout,
			Logger:      discardLogger{},
		}, nil
	}
	return c.hostClient()
}

func (c *Config) hostClient() (*fasthttp.HostClient, error) {
	hc := &fasthttp.HostClient{
		Name:        "httpgo/" + Version,
		Addr:        c.addr,
		Dial:        c.getDialer(),
		IsTLS:       c.isTLS,
		TLSConfig:   c.tlsConf,
		MaxConns:    c.Connections,
		ReadTimeout: c.Timeout,
	}

	if c.Http2 {
		log.Println("由于兼容性问题，HTTP/2 支持已被暂时禁用")
		// TODO: 使用兼容库重新启用 HTTP/2 支持
		// if err := http2.ConfigureClient(hc, http2.ClientOpts{}); err != nil {
		//     return nil, fmt.Errorf("%s doesn't support http/2\n", hc.Addr)
		// }
	}

	return hc, nil
}

func (c *Config) setReqBasic(req *fasthttp.Request) (err error) {
	req.Header.SetMethod(c.Method)
	req.SetRequestURI(c.Url)

	uri := req.URI()
	host := uri.Host()

	scheme := uri.Scheme()
	if bytes.Equal(scheme, strHTTPS) {
		c.isTLS = true
	} else if !bytes.Equal(scheme, strHTTP) {
		err = fmt.Errorf("unsupported protocol %q. http and https are supported", scheme)
		return
	}

	c.addr = addMissingPort(string(host), c.isTLS)

	return
}

var (
	strHTTP  = []byte("http")
	strHTTPS = []byte("https")
)

func addMissingPort(addr string, isTLS bool) string {
	n := strings.Index(addr, ":")
	if n >= 0 {
		return addr
	}
	port := 80
	if isTLS {
		port = 443
	}
	return net.JoinHostPort(addr, strconv.Itoa(port))
}

func (c *Config) setReqBody(req *fasthttp.Request) (err error) {
	if c.Body != "" {
		c.body = []byte(c.Body)
	}

	if c.File != "" {
		c.body, err = os.ReadFile(filepath.Clean(c.File))
	}

	if !c.Stream {
		// 设置固定请求体
		req.SetBody(c.body)
	}

	return
}

// parseArgs 从额外参数中获取请求体
func (c *Config) parseArgs() {
	if len(c.Args) == 0 {
		return
	}

	isJson := true
	for _, arg := range c.Args {
		// 检查参数格式有效性
		if arg == "" {
			continue
		}

		formEqualIndex := strings.Index(arg, "=")
		jsonEqualIndex := strings.Index(arg, ":=")
		// 没有 "=" 或 "=" 在 ":=" 之前
		if formEqualIndex == -1 || jsonEqualIndex == -1 || formEqualIndex < jsonEqualIndex {
			isJson = false
			break
		}
	}

	if isJson {
		c.buildJSONBody()
	} else {
		c.buildFormBody()
	}
}

// buildJSONBody 构建 JSON 请求体
func (c *Config) buildJSONBody() {
	c.JSON = true
	c.body = append(c.body, '{')
	for ii, arg := range c.Args {
		if arg == "" {
			continue
		}

		i := strings.Index(arg, ":=")
		if i == -1 {
			continue // 跳过无效参数
		}

		k, v := strings.TrimSpace(arg[:i]), strings.TrimSpace(arg[i+2:])
		if k == "" {
			continue // 跳过空键名
		}

		c.body = append(c.body, '"')
		c.body = append(c.body, k...)
		c.body = append(c.body, '"', ':')

		if needQuote(v) {
			c.body = append(c.body, '"')
			c.body = append(c.body, v...)
			c.body = append(c.body, '"')
		} else {
			c.body = append(c.body, v...)
		}

		if ii < len(c.Args)-1 {
			c.body = append(c.body, ',')
		}
	}
	c.body = append(c.body, '}')
}

// buildFormBody 构建表单请求体
func (c *Config) buildFormBody() {
	c.Form = true
	c.Method = fasthttp.MethodPost
	formArgs := fasthttp.AcquireArgs()
	defer fasthttp.ReleaseArgs(formArgs)

	for _, arg := range c.Args {
		if arg == "" {
			continue
		}

		i := strings.Index(arg, "=")
		if i == -1 {
			formArgs.AddNoValue(strings.TrimSpace(arg))
		} else {
			key := strings.TrimSpace(arg[:i])
			value := strings.TrimSpace(arg[i+1:])
			if key != "" {
				formArgs.Add(key, value)
			}
		}
	}
	c.body = formArgs.AppendBytes(c.body)
}

func needQuote(v string) bool {
	if vv := strings.ToLower(v); vv == "false" || vv == "true" {
		return false
	}
	if _, err := strconv.Atoi(v); err == nil {
		return false
	}
	if _, err := strconv.ParseFloat(v, 64); err == nil {
		return false
	}

	l := len(v)
	if l <= 1 {
		return true
	}

	if (v[0] == '[' && v[l-1] == ']') || (v[0] == '{' && v[l-1] == '}') {
		return false
	}

	return true
}

func (c *Config) setReqHeader(req *fasthttp.Request) (err error) {
	if err = headers(c.Headers).writeToHttp(req); err != nil {
		return
	}

	if c.DisableKeepAlives {
		req.Header.SetConnectionClose()
	}
	if c.Host != "" {
		req.URI().SetHost(c.Host)
	}

	if c.JSON {
		req.Header.SetContentType(mimeApplicationJSON)
	}
	if c.Form {
		req.Header.SetContentType(mimeApplicationForm)
	}

	return
}

func (c *Config) getDialer() fasthttp.DialFunc {
	if c.HttpProxy != "" {
		return httpProxyDialer(&c.throughput, c.HttpProxy, c.Timeout)
	}
	if c.SocksProxy != "" {
		return httpSocksProxyDialer(&c.throughput, c.SocksProxy)
	}

	return httpDialer(&c.throughput, c.Timeout)
}

/* #nosec G402 */
func (c *Config) getTlsConfig() (conf *tls.Config, err error) {
	var certs []tls.Certificate
	if certs, err = readClientCert(c.Cert, c.Key); err != nil {
		return
	}
	conf = &tls.Config{
		Certificates:       certs,
		InsecureSkipVerify: c.Insecure, // 允许不安全的连接
	}

	return
}

func readClientCert(certPath, keyPath string) (certs []tls.Certificate, err error) {
	if certPath == "" && keyPath == "" {
		return
	}

	var cert tls.Certificate
	if cert, err = tls.LoadX509KeyPair(certPath, keyPath); err != nil {
		return
	}

	certs = append(certs, cert)

	return
}

func (c *Config) getMaxRedirects() int {
	if !c.Follow {
		return 0
	}

	n := c.MaxRedirects
	if n <= 0 {
		n = defaultMaxRedirects
	}

	return n
}
