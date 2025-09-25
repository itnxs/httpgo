package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/itnxs/httpgo/internal/pkg"
	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}

var config pkg.Config

func init() {
	rootCmd.Flags().SortFlags = false
	rootCmd.Flags().IntVarP(&config.Connections, "connections", "c", 128, "最大并发连接数")
	rootCmd.Flags().IntVarP(&config.Count, "requests", "n", 0, "请求总数（如果指定，则忽略 --duration 参数）")
	rootCmd.Flags().IntVar(&config.Qps, "qps", 0, "固定基准测试的最大 QPS 值（如果指定，则忽略 -n|--requests 参数）")
	rootCmd.Flags().DurationVarP(&config.Duration, "duration", "d", time.Second*10, "测试持续时间")
	rootCmd.Flags().DurationVarP(&config.Timeout, "timeout", "t", time.Second*3, "请求超时时间")
	rootCmd.Flags().StringVarP(&config.Method, "method", "X", "GET", "HTTP 请求方法")
	rootCmd.Flags().StringSliceVarP(&config.Headers, "header", "H", nil, headersUsage)
	rootCmd.Flags().StringVar(&config.Host, "host", "", "覆盖请求主机名")
	rootCmd.Flags().BoolVarP(&config.DisableKeepAlives, "disableKeepAlives", "a", false, "禁用 HTTP keep-alive，如果为 true，将设置 Connection: close 头")
	rootCmd.Flags().StringVarP(&config.Body, "body", "b", "", "HTTP 请求体字符串")
	rootCmd.Flags().StringVarP(&config.File, "file", "f", "", "从文件路径读取 HTTP 请求体")
	rootCmd.Flags().BoolVarP(&config.Stream, "stream", "s", false, "使用流式请求体以减少内存使用")
	rootCmd.Flags().BoolVarP(&config.JSON, "json", "J", false, "发送 JSON 请求，自动设置 Content-Type 为 application/json")
	rootCmd.Flags().BoolVarP(&config.Form, "form", "F", false, "发送表单请求，自动设置 Content-Type 为 application/x-www-form-urlencoded")
	rootCmd.Flags().BoolVarP(&config.Insecure, "insecure", "k", false, "控制客户端是否验证服务器的证书链和主机名")
	rootCmd.Flags().StringVar(&config.Cert, "cert", "", "客户端 TLS 证书路径")
	rootCmd.Flags().StringVar(&config.Key, "key", "", "客户端 TLS 证书私钥路径")
	rootCmd.Flags().StringVar(&config.HttpProxy, "httpProxy", "", "HTTP 代理地址")
	rootCmd.Flags().StringVar(&config.SocksProxy, "socksProxy", "", "SOCKS 代理地址")
	rootCmd.Flags().BoolVarP(&config.Pipeline, "pipeline", "p", false, "使用 fasthttp 管道客户端")
	rootCmd.Flags().BoolVar(&config.Follow, "follow", false, "在调试模式下跟随 30x 重定向")
	rootCmd.Flags().IntVar(&config.MaxRedirects, "maxRedirects", 0, "跟随 30x 重定向的最大次数，默认为 30（配合 --follow 使用）")
	rootCmd.Flags().BoolVarP(&config.Debug, "debug", "D", false, "只发送一次请求并显示请求和响应详情")
	rootCmd.Flags().BoolVar(&config.Http2, "http2", false, "使用 HTTP/2.0")
}

var rootCmd = &cobra.Command{
	Use:           usage,
	Example:       example,
	Short:         "httpgo 是一个快速的 HTTP 基准测试工具",
	Version:       pkg.Version,
	Args:          rootArgs,
	Run:           rootRun,
	SilenceErrors: true,
}

func rootArgs(_ *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("缺少 URL 参数")
	}
	return nil
}

func rootRun(cmd *cobra.Command, args []string) {
	config.Url = args[0]
	config.Args = args[1:]
	if err := pkg.New(config).Run(); err != nil {
		cmd.PrintErrln(err)
	}
}

const (
	usage   = `httpgo [url|:port|/path] [k:v|k:=v ...]`
	example = `	httpgo https://www.google.com -c1 -n5   =>   httpgo -X GET https://www.google.com -c1 -n5
	httpgo :3000 -c1 -n5                    =>   httpgo -X GET http://localhost:3000 -c1 -n5
	httpgo /foo -c1 -n5                     =>   httpgo -X GET http://localhost/foo -c1 -n5
	httpgo :3000 -c1 -n5 foo:=bar           =>   httpgo -X GET http://localhost:3000 -c1 -n5 -H "Content-Type: application/json" -b='{"foo":"bar"}'
	httpgo :3000 -c1 -n5 foo=bar            =>   httpgo -X POST http://localhost:3000 -c1 -n5 -H "Content-Type: application/x-www-form-urlencoded" -b="foo=bar"`
	headersUsage = `格式为 "K: V" 的 HTTP 请求头，可重复使用
示例:
	-H "k1: v1" -H k2:v2
	-H "k3: v3, k4: v4"`
)
