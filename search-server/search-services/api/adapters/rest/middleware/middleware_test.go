package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRateMiddleware(t *testing.T) {
	t.Parallel()
	var count int64

	f := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&count, 1)
	})

	handler := Rate(f, 2)

	start := time.Now()

	for range 10 {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	delta := time.Since(start)

	require.Equal(t, int64(10), atomic.LoadInt64(&count))
	require.GreaterOrEqual(t, delta, 4*time.Second)
}

func TestBadArgumentRate(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	require.PanicsWithValue(t, "rpc limit must be positive", func() {
		Rate(handler, 0)
	})
}

func TestConcurrency(t *testing.T) {
	t.Parallel()

	f := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	})

	handler := Concurrency(f, 100)

	var count int64

	wg := new(sync.WaitGroup)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	for range 200 {
		rec := httptest.NewRecorder()
		wg.Go(func() {
			handler.ServeHTTP(rec, req)
			if rec.Code == http.StatusServiceUnavailable {
				atomic.AddInt64(&count, 1)
			}
		})
	}
	wg.Wait()

	require.GreaterOrEqual(t, count, int64(90))
}

func TestBadArgumentConcurrency(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	require.PanicsWithValue(t, "concurrency limit must be positive", func() {
		Concurrency(handler, 0)
	})
}

// type fakeVerifier struct {
// 	verifyFn func(token string) error
// }

// func (f fakeVerifier) Verify(token string) error {
// 	return f.verifyFn(token)
// }

// func TestAuth(t *testing.T) {
// 	t.Parallel()

// 	tests := []struct {
// 		name           string
// 		authHeader     string
// 		verifyFn       func(token string) error
// 		wantStatusCode int
// 		wantCalledNext bool
// 	}{
// 		{
// 			name:           "NO_HEADER",
// 			authHeader:     "",
// 			verifyFn:       func(token string) error { return nil },
// 			wantStatusCode: http.StatusUnauthorized,
// 			wantCalledNext: false,
// 		},
// 		{
// 			name:           "WRONG_PREFIX",
// 			authHeader:     "Golang abc",
// 			verifyFn:       func(token string) error { return nil },
// 			wantStatusCode: http.StatusUnauthorized,
// 			wantCalledNext: false,
// 		},
// 		{
// 			name:           "EMPTY_TOKEN",
// 			authHeader:     "Token ",
// 			verifyFn:       func(token string) error { return nil },
// 			wantStatusCode: http.StatusUnauthorized,
// 			wantCalledNext: false,
// 		},
// 		{
// 			name:       "VERIFY_ERROR",
// 			authHeader: "Token abc",
// 			verifyFn: func(token string) error {
// 				require.Equal(t, "abc", token)
// 				return errors.New("bad token")
// 			},
// 			wantStatusCode: http.StatusUnauthorized,
// 			wantCalledNext: false,
// 		},
// 		{
// 			name:       "OK",
// 			authHeader: "Token abc",
// 			verifyFn: func(token string) error {
// 				require.Equal(t, "abc", token)
// 				return nil
// 			},
// 			wantStatusCode: http.StatusOK,
// 			wantCalledNext: true,
// 		},
// 	}

// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			t.Parallel()

// 			calledNext := false

// 			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 				calledNext = true
// 			})

// 			handler := Auth(next, fakeVerifier{
// 				verifyFn: tc.verifyFn,
// 			})

// 			req := httptest.NewRequest(http.MethodGet, "/", nil)
// 			if tc.authHeader != "" {
// 				req.Header.Set("Authorization", tc.authHeader)
// 			}

// 			rec := httptest.NewRecorder()
// 			handler.ServeHTTP(rec, req)

// 			require.Equal(t, tc.wantStatusCode, rec.Code)
// 			require.Equal(t, tc.wantCalledNext, calledNext)
// 		})
// 	}
// }
