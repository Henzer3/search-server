package middleware

import (
	"net/http"

	"go.uber.org/ratelimit"
)

func Rate(next http.Handler, rps int) http.Handler {
	if rps <= 0 {
		panic("rpc limit must be positive")
	}

	rateLimiter := ratelimit.New(rps)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rateLimiter.Take()
		next.ServeHTTP(w, r)
	})
}
