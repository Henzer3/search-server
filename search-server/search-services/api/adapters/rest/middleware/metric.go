package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/VictoriaMetrics/metrics"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		recorder := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(recorder, r)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(recorder.status)
		url := r.URL.Path

		metrics.GetOrCreateHistogram(
			`http_request_duration_seconds{status="` + status + `",url="` + url + `"}`,
		).Update(duration)
	})
}
