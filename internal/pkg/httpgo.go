package pkg

import (
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Version 信息，可在构建时通过 ldflags 注入
var (
	Version   = "1.0.0"
	Commit    = "unknown"
	BuildTime = "unknown"
)

const (
	// 使用常量文件中的默认值
	defaultConnections  = DefaultConnections
	defaultDuration     = DefaultDuration
	defaultTimeout      = DefaultTimeout
	defaultMaxRedirects = DefaultMaxRedirects
)

// HttpGo denotes httpgo application
type HttpGo struct {
	c *Config
	client
	limiter
	wg sync.WaitGroup

	mut       sync.Mutex
	startTime time.Time
	roundReqs int64
	done      bool
	doneChan  chan struct{}
	*stat
}

// New create a HttpGo instance with specific Config
func New(c Config) *HttpGo {
	p := &HttpGo{
		c:        &c,
		limiter:  nopeLimiter(false),
		doneChan: make(chan struct{}),
	}

	if p.c.Connections <= 0 {
		p.c.Connections = defaultConnections
	}

	if p.c.Duration <= 0 {
		p.c.Duration = defaultDuration
	}

	if p.c.Timeout <= 0 {
		p.c.Timeout = defaultTimeout
	}

	if p.c.MaxRedirects <= 0 {
		p.c.MaxRedirects = defaultMaxRedirects
	}

	p.stat = newStat()
	p.stat.count = p.c.Count
	p.stat.duration = p.c.Duration
	p.stat.connections = p.c.Connections
	p.stat.throughput = &p.c.throughput
	p.initCmd = p.run

	return p
}

// Run starts benchmarking
func (p *HttpGo) Run() (err error) {
	if err = p.init(); err != nil {
		return
	}

	if p.c.Debug {
		return p.doOnce()
	}

	return p.stat.start()
}

func (p *HttpGo) init() (err error) {
	if p.c.Url == "" {
		return errors.New("缺少 URL 参数")
	}

	p.c.Url = addMissingSchemaAndHost(p.c.Url)
	p.stat.url = p.c.Url

	if p.c.Qps > 0 {
		p.limiter = newTokenLimiter(p.c.Qps)
	}

	if p.client == nil {
		p.client, err = newHttpClient(p.c)
	}

	return
}

func addMissingSchemaAndHost(url string) string {
	if !strings.HasPrefix(url, "://") && strings.HasPrefix(url, ":") {
		return "http://localhost" + url
	}
	if strings.Index(url, "://") == -1 && len(url) >= 2 {
		if url[0] == '/' && url[1] != '/' {
			return "http://localhost" + url
		}
		if url[0] != '/' && url[1] != '/' {
			return "http://" + url
		}
	}
	return url
}

func (p *HttpGo) run() tea.Msg {
	p.startTime = time.Now()
	n := p.c.Connections
	p.wg.Add(n)
	for i := 0; i < n; i++ {
		go p.worker()
	}
	p.wg.Wait()

	return done
}

func (p *HttpGo) worker() {
	defer p.wg.Done()

	for {
		select {
		case <-p.doneChan:
			return
		default:
			if p.limiter.allow() {
				p.statistic(p.do())
			}
		}
	}
}

const interval = time.Millisecond * 10

func (p *HttpGo) statistic(code int, latency time.Duration, err error) {
	p.mut.Lock()
	defer p.mut.Unlock()
	if p.done {
		return
	}

	if err != nil {
		p.appendError(err)
	} else {
		p.roundReqs++
		atomic.AddInt64(&p.reqs, 1)
		p.appendCode(code)
		p.appendLatency(latency)
	}

	elapsed := time.Since(p.startTime)
	if p.c.Count > 0 && atomic.LoadInt64(&p.reqs) == int64(p.c.Count) {
		p.appendRps(float64(p.roundReqs) / elapsed.Seconds())
		p.done = true
		close(p.doneChan)
		return
	}

	if elapsed >= interval {
		p.appendRps(float64(p.roundReqs) / elapsed.Seconds())

		atomic.AddInt64(&p.elapsed, int64(elapsed))

		p.startTime = time.Now()
		p.roundReqs = 0
	}

	if p.c.Count <= 0 && atomic.LoadInt64(&p.elapsed) >= int64(p.c.Duration) {
		p.done = true
		close(p.doneChan)
	}
}
