package pkg

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

type client interface {
	do() (int, time.Duration, error)
	doOnce() error
}

type clientDoer interface {
	Do(*fasthttp.Request, *fasthttp.Response) error
}

type onceClientDoer interface {
	clientDoer
	DoRedirects(*fasthttp.Request, *fasthttp.Response, int) error
}

type httpClient struct {
	doer         clientDoer
	onceDoer     onceClientDoer
	requestPool  sync.Pool
	streamPool   sync.Pool
	request      *fasthttp.Request
	writeCloser  io.WriteCloser
	body         []byte
	stream       bool
	maxRedirects int
}

func newHttpClient(c *Config) (fc *httpClient, err error) {
	fc = &httpClient{
		maxRedirects: c.getMaxRedirects(),
		request:      fasthttp.AcquireRequest(),
		stream:       c.Stream,
		writeCloser:  defaultWriteCloser{Writer: os.Stdout},
	}

	c.parseArgs()
	if err = c.setReqBasic(fc.request); err != nil {
		return
	}
	if err = c.setReqBody(fc.request); err != nil {
		return
	}
	fc.body = c.body
	if err = c.setReqHeader(fc.request); err != nil {
		return
	}
	if c.tlsConf, err = c.getTlsConfig(); err != nil {
		return
	}

	if c.Debug {
		fc.request.SetConnectionClose()
		fc.onceDoer, err = c.hostClient()
	} else {
		fc.doer, err = c.doer()
	}

	return
}

func (c *httpClient) do() (code int, latency time.Duration, err error) {
	var (
		req  = c.acquireReq()
		resp = fasthttp.AcquireResponse()
	)

	defer func() {
		c.requestPool.Put(req)
		fasthttp.ReleaseResponse(resp)
	}()

	if c.stream {
		bodyStream := c.acquireBodyStream()
		req.SetBodyStream(bodyStream, -1)
		c.streamPool.Put(bodyStream)
	}

	start := time.Now()
	if err = c.doer.Do(req, resp); err != nil {
		return
	}

	code = resp.StatusCode()
	latency = time.Since(start)

	return
}

func (c *httpClient) acquireReq() *fasthttp.Request {
	v := c.requestPool.Get()
	if v == nil {
		req := fasthttp.AcquireRequest()
		c.request.CopyTo(req)
		return req
	}

	return v.(*fasthttp.Request)
}

func (c *httpClient) acquireBodyStream() *bytes.Reader {
	v := c.streamPool.Get()
	if v == nil {
		return bytes.NewReader(c.body)
	}
	bodyStream := v.(*bytes.Reader)
	bodyStream.Reset(c.body)
	return bodyStream
}

func (c *httpClient) doOnce() (err error) {
	var (
		req  = c.request
		resp = fasthttp.AcquireResponse()
	)

	if c.stream {
		req.SetBodyStream(bytes.NewReader(c.body), -1)
	}

	defer func() {
		if err == nil {
			msg := fmt.Sprintf("Connected to %s(%v)\r\n\r\n", req.URI().Host(), resp.RemoteAddr())
			_, _ = c.writeCloser.Write([]byte(msg))
			_, _ = req.WriteTo(c.writeCloser)
			_, _ = c.writeCloser.Write([]byte("\n\n"))
			_, _ = resp.WriteTo(c.writeCloser)
			_ = c.writeCloser.Close()
		}
	}()

	if c.maxRedirects > 0 {
		err = c.onceDoer.DoRedirects(c.request, resp, c.maxRedirects)
	} else {
		err = c.onceDoer.Do(c.request, resp)
	}

	return
}

type discardLogger struct{}

func (discardLogger) Printf(_ string, _ ...interface{}) {}

type defaultWriteCloser struct {
	io.Writer
}

func (wc defaultWriteCloser) Close() error {
	return nil
}
