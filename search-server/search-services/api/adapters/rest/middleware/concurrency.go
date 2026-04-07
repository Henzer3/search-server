package middleware

import (
	"net/http"
)

type concurrencyLimiter struct {
	ch chan struct{}
}

func newLimiter(thread int) *concurrencyLimiter {
	return &concurrencyLimiter{ch: make(chan struct{}, thread)}
}

func (c *concurrencyLimiter) allow() bool {
	select {
	case c.ch <- struct{}{}:
		return true
	default:
		return false
	}
}

func (c *concurrencyLimiter) release() {
	select {
	case <-c.ch:
	default:
	}
}

func Concurrency(next http.Handler, limit int) http.Handler {
	if limit <= 0 {
		panic(" concurrency limit must be positive")
	}

	cl := newLimiter(limit)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cl.allow() {
			defer cl.release()
			next.ServeHTTP(w, r)
			return
		}
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
	})
}
