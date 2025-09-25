package pkg

import (
	"sync"
	"time"
)

type limiter interface {
	allow() bool
}

type nopeLimiter bool

func (nopeLimiter) allow() bool { return true }

type tokenLimiter struct {
	mut   sync.Mutex
	limit float64
	burst int
	token int
	last  time.Time
}

func newTokenLimiter(qps int) *tokenLimiter {
	return &tokenLimiter{
		limit: float64(qps),
		burst: qps,
		token: qps,
		last:  time.Now(),
	}
}

func (t *tokenLimiter) allow() bool {
	t.mut.Lock()
	defer t.mut.Unlock()

	if t.token -= t.revoked(time.Since(t.last)); t.token < 0 {
		t.token = 0
	}

	if t.token < t.burst {
		t.token++
		t.last = time.Now()
		return true
	}

	return false
}

func (t *tokenLimiter) revoked(d time.Duration) int {
	sec := float64(d/time.Second) * t.limit
	nsec := float64(d%time.Second) * t.limit
	return int(sec + nsec/1e9)
}
