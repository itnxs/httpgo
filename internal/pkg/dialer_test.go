package pkg

import (
    "net"
    "runtime"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/valyala/fasthttp"
)

func Test_httpDialer(t *testing.T) {
    t.Parallel()

    ln, err := net.Listen("tcp", "127.0.0.1:0")
    assert.Nil(t, err)

    addr := ln.Addr().String()

    go func() {
        assert.Nil(t, fasthttp.Serve(ln, func(_ *fasthttp.RequestCtx) {}))
    }()

    var throughput int64

    t.Run("timeout", func(t *testing.T) {
        if runtime.GOOS == "windows" {
            t.Skip("skip windows test")
        }

        hc := &fasthttp.HostClient{
            Addr: addr,
            Dial: httpDialer(&throughput, time.Nanosecond),
        }

        req := &fasthttp.Request{}
        req.SetRequestURI(addr)
        req.URI().SetHost("127.0.0.1")
        resp := &fasthttp.Response{}

        err = hc.Do(req, resp)
        assert.NotNil(t, err)
        assert.Contains(t, err.Error(), "timed out")
    })

    t.Run("success", func(t *testing.T) {
        hc := &fasthttp.HostClient{
            Addr: addr,
            Dial: httpDialer(&throughput, time.Second*3),
        }

        req := &fasthttp.Request{}
        req.SetRequestURI(addr)
        req.URI().SetHost("127.0.0.1")
        resp := &fasthttp.Response{}

        assert.Nil(t, hc.Do(req, resp))
        assert.Equal(t, int64(165), throughput)
    })
}

func Test_httpHttpProxyDialer(t *testing.T) {
    t.Parallel()

    ln, err := net.Listen("tcp", "127.0.0.1:0")
    assert.Nil(t, err)

    addr := ln.Addr().String()

    go func() {
        assert.Nil(t, fasthttp.Serve(ln, func(ctx *fasthttp.RequestCtx) {}))
    }()

    var throughput int64

    t.Run("timeout", func(t *testing.T) {
        if runtime.GOOS == "windows" {
            t.Skip("skip windows test")
        }
        hc := &fasthttp.HostClient{
            Addr: addr,
            Dial: httpProxyDialer(&throughput, addr, time.Nanosecond),
        }

        req := &fasthttp.Request{}
        req.SetRequestURI(addr)
        req.URI().SetHost("127.0.0.1")
        resp := &fasthttp.Response{}

        err = hc.Do(req, resp)
        assert.NotNil(t, err)
        assert.Contains(t, err.Error(), "timed out")
    })

    t.Run("0 timeout", func(t *testing.T) {
        hc := &fasthttp.HostClient{
            Addr: addr,
            Dial: httpProxyDialer(&throughput, addr, 0),
        }

        req := &fasthttp.Request{}
        req.SetRequestURI(addr)
        req.URI().SetHost("127.0.0.1")
        resp := &fasthttp.Response{}

        err = hc.Do(req, resp)
        assert.Nil(t, err)
    })

    t.Run("success", func(t *testing.T) {
        hc := &fasthttp.HostClient{
            Addr: addr,
            Dial: httpProxyDialer(&throughput, "a:b@"+addr, time.Second*3),
        }

        req := &fasthttp.Request{}
        req.SetRequestURI(addr)
        req.URI().SetHost("127.0.0.1")
        resp := &fasthttp.Response{}

        assert.Nil(t, hc.Do(req, resp))
        assert.True(t, throughput > 0)
    })
}

func Test_httpSocksProxyDialer(t *testing.T) {
    t.Parallel()

    ln, err := net.Listen("tcp", "127.0.0.1:0")
    assert.Nil(t, err)

    addr := ln.Addr().String()

    go func() {
        assert.Nil(t, fasthttp.Serve(ln, func(ctx *fasthttp.RequestCtx) {}))
    }()

    var throughput int64

    t.Run("error proxy", func(t *testing.T) {
        hc := &fasthttp.HostClient{
            Addr: addr,
            Dial: httpSocksProxyDialer(&throughput, addr),
        }

        req := &fasthttp.Request{}
        req.SetRequestURI(addr)
        req.URI().SetHost("127.0.0.1")
        resp := &fasthttp.Response{}

        err = hc.Do(req, resp)
        assert.NotNil(t, err)
        assert.Contains(t, err.Error(), "socks proxy")
    })

    t.Run("success", func(t *testing.T) {
        hc := &fasthttp.HostClient{
            Addr: addr,
            Dial: httpSocksProxyDialer(&throughput, "socks5://127.0.0.1:88888"),
        }

        req := &fasthttp.Request{}
        req.SetRequestURI(addr)
        req.URI().SetHost("127.0.0.1")
        resp := &fasthttp.Response{}

        assert.NotNil(t, hc.Do(req, resp))
    })
}
