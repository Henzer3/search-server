package rest

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"yadro.com/course/api/core"
	"yadro.com/course/api/core/mocks"
)

var logger = slog.New(slog.NewTextHandler(io.Discard, nil))

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func TestWordsHandler(t *testing.T) {
	tests := []struct {
		name       string
		phrase     string
		mockSetup  func(norm *mocks.MockNormalizer)
		wantStatus int
	}{
		{
			name:   "OK",
			phrase: "I love golang so much",
			mockSetup: func(norm *mocks.MockNormalizer) {
				norm.EXPECT().Norm(gomock.Any(), "I love golang so much").Return([]string{"love", "golang"}, nil).Times(1)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "TooLargeMessage",
			phrase: strings.Repeat("s", 4096+1),
			mockSetup: func(norm *mocks.MockNormalizer) {
				norm.EXPECT().Norm(gomock.Any(), strings.Repeat("s", 4096+1)).Return(nil, core.ErrLimit).Times(1)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "AnotherError",
			phrase: "golang better then C++",
			mockSetup: func(norm *mocks.MockNormalizer) {
				norm.EXPECT().Norm(gomock.Any(), "golang better then C++").Return(nil, errors.New("internal error")).Times(1)
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:   "EmptyPhrase",
			phrase: "",
			mockSetup: func(norm *mocks.MockNormalizer) {
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			norm := mocks.NewMockNormalizer(ctrl)
			tc.mockSetup(norm)
			f := NewWordsHandler(logger, norm)

			w := httptest.NewRecorder()
			recorder := &statusRecorder{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			req := httptest.NewRequest(http.MethodPost, "/", nil)
			q := req.URL.Query()
			q.Set("phrase", tc.phrase)
			req.URL.RawQuery = q.Encode()
			f(recorder, req)
			require.Equal(t, tc.wantStatus, recorder.status)
		})

	}
}

func TestUpdateHandler(t *testing.T) {
	tests := []struct {
		name       string
		mockSetup  func(updater *mocks.MockUpdater)
		wantStatus int
	}{
		{
			name: "OK",
			mockSetup: func(updater *mocks.MockUpdater) {
				updater.EXPECT().Update(gomock.Any()).Return(nil).Times(1)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "AlreadyUpdating",
			mockSetup: func(updater *mocks.MockUpdater) {
				updater.EXPECT().Update(gomock.Any()).Return(core.ErrAlreadyUpdating).Times(1)
			},
			wantStatus: http.StatusAccepted,
		},
		{
			name: "InternalError",
			mockSetup: func(updater *mocks.MockUpdater) {
				updater.EXPECT().Update(gomock.Any()).Return(errors.New("internal error")).Times(1)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			updater := mocks.NewMockUpdater(ctrl)
			tc.mockSetup(updater)

			f := NewUpdateHandler(logger, updater)

			w := httptest.NewRecorder()
			recorder := &statusRecorder{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			req := httptest.NewRequest(http.MethodPost, "/", nil)
			f(recorder, req)

			require.Equal(t, tc.wantStatus, recorder.status)
		})
	}
}

func TestUpdateStatsHandler(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(updater *mocks.MockUpdater)
		wantStatus    int
		wantBody      StatsResponse
		wantBodyCheck bool
	}{
		{
			name: "OK",
			mockSetup: func(updater *mocks.MockUpdater) {
				updater.EXPECT().Stats(gomock.Any()).Return(core.UpdateStats{
					WordsTotal:    100,
					WordsUnique:   50,
					ComicsFetched: 10,
					ComicsTotal:   200,
				}, nil).Times(1)
			},
			wantStatus: http.StatusOK,
			wantBody: StatsResponse{
				WordsTotal:    100,
				WordsUnique:   50,
				ComicsFetched: 10,
				ComicsTotal:   200,
			},
			wantBodyCheck: true,
		},
		{
			name: "InternalError",
			mockSetup: func(updater *mocks.MockUpdater) {
				updater.EXPECT().Stats(gomock.Any()).Return(core.UpdateStats{}, errors.New("stats error")).Times(1)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			updater := mocks.NewMockUpdater(ctrl)
			tc.mockSetup(updater)

			f := NewUpdateStatsHandler(logger, updater)

			w := httptest.NewRecorder()
			recorder := &statusRecorder{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			f(recorder, req)

			require.Equal(t, tc.wantStatus, recorder.status)

			if tc.wantBodyCheck {
				require.Equal(t, "application/json", w.Header().Get("Content-Type"))

				var res StatsResponse
				err := json.NewDecoder(w.Body).Decode(&res)
				require.NoError(t, err)
				require.Equal(t, tc.wantBody, res)
			}
		})
	}
}

func TestUpdateStatusHandler(t *testing.T) {
	tests := []struct {
		name       string
		mockSetup  func(updater *mocks.MockUpdater)
		wantStatus int
		wantBody   UpdateStatus
	}{
		{
			name: "Idle",
			mockSetup: func(updater *mocks.MockUpdater) {
				updater.EXPECT().Status(gomock.Any()).Return(core.StatusUpdateIdle, nil).Times(1)
			},
			wantStatus: http.StatusOK,
			wantBody:   UpdateStatus{Status: "idle"},
		},
		{
			name: "Running",
			mockSetup: func(updater *mocks.MockUpdater) {
				updater.EXPECT().Status(gomock.Any()).Return(core.StatusUpdateRunning, nil).Times(1)
			},
			wantStatus: http.StatusOK,
			wantBody:   UpdateStatus{Status: "running"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			updater := mocks.NewMockUpdater(ctrl)
			tc.mockSetup(updater)

			f := NewUpdateStatusHandler(logger, updater)

			w := httptest.NewRecorder()
			recorder := &statusRecorder{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			f(recorder, req)

			require.Equal(t, tc.wantStatus, recorder.status)
			require.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var res UpdateStatus
			err := json.NewDecoder(w.Body).Decode(&res)
			require.NoError(t, err)
			require.Equal(t, tc.wantBody, res)
		})
	}
}

func TestDropHandler(t *testing.T) {
	tests := []struct {
		name       string
		mockSetup  func(updater *mocks.MockUpdater)
		wantStatus int
	}{
		{
			name: "OK",
			mockSetup: func(updater *mocks.MockUpdater) {
				updater.EXPECT().Drop(gomock.Any()).Return(nil).Times(1)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "InternalError",
			mockSetup: func(updater *mocks.MockUpdater) {
				updater.EXPECT().Drop(gomock.Any()).Return(errors.New("drop error")).Times(1)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			updater := mocks.NewMockUpdater(ctrl)
			tc.mockSetup(updater)

			f := NewDropHandler(logger, updater)

			w := httptest.NewRecorder()
			recorder := &statusRecorder{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			req := httptest.NewRequest(http.MethodDelete, "/", nil)
			f(recorder, req)

			require.Equal(t, tc.wantStatus, recorder.status)
		})
	}
}

func TestSearchHandler(t *testing.T) {
	tests := []struct {
		name          string
		phrase        string
		limit         string
		mockSetup     func(searcher *mocks.MockSearcher)
		wantStatus    int
		wantBody      SearchResponse
		wantBodyCheck bool
	}{
		{
			name:   "OK",
			phrase: "golang",
			limit:  "2",
			mockSetup: func(searcher *mocks.MockSearcher) {
				searcher.EXPECT().
					Search(gomock.Any(), "golang", 2).
					Return([]core.ImageInformation{
						{ID: 1, Url: "https://example.com/1.jpg"},
						{ID: 2, Url: "https://example.com/2.jpg"},
					}, nil).
					Times(1)
			},
			wantStatus: http.StatusOK,
			wantBody: SearchResponse{
				Comics: []ComicResponse{
					{ID: 1, URL: "https://example.com/1.jpg"},
					{ID: 2, URL: "https://example.com/2.jpg"},
				},
				Total: 2,
			},
			wantBodyCheck: true,
		},
		{
			name:   "OkDefaultLimit",
			phrase: "docker",
			limit:  "",
			mockSetup: func(searcher *mocks.MockSearcher) {
				searcher.EXPECT().
					Search(gomock.Any(), "docker", core.DefaultLimitValue).
					Return([]core.ImageInformation{
						{ID: 10, Url: "https://example.com/10.jpg"},
					}, nil).
					Times(1)
			},
			wantStatus: http.StatusOK,
			wantBody: SearchResponse{
				Comics: []ComicResponse{
					{ID: 10, URL: "https://example.com/10.jpg"},
				},
				Total: 1,
			},
			wantBodyCheck: true,
		},
		{
			name:          "EmptyPhrase",
			phrase:        "",
			limit:         "2",
			mockSetup:     func(searcher *mocks.MockSearcher) {},
			wantStatus:    http.StatusBadRequest,
			wantBodyCheck: false,
		},
		{
			name:          "BadLimit",
			phrase:        "golang",
			limit:         "abc",
			mockSetup:     func(searcher *mocks.MockSearcher) {},
			wantStatus:    http.StatusBadRequest,
			wantBodyCheck: false,
		},
		{
			name:          "LimitZero",
			phrase:        "golang",
			limit:         "0",
			mockSetup:     func(searcher *mocks.MockSearcher) {},
			wantStatus:    http.StatusBadRequest,
			wantBodyCheck: false,
		},
		{
			name:   "InternalError",
			phrase: "golang",
			limit:  "3",
			mockSetup: func(searcher *mocks.MockSearcher) {
				searcher.EXPECT().
					Search(gomock.Any(), "golang", 3).
					Return(nil, errors.New("internal error")).
					Times(1)
			},
			wantStatus:    http.StatusInternalServerError,
			wantBodyCheck: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			searcher := mocks.NewMockSearcher(ctrl)
			tc.mockSetup(searcher)

			f := NewSearchHandler(logger, searcher)

			w := httptest.NewRecorder()
			recorder := &statusRecorder{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			q := req.URL.Query()
			if tc.phrase != "" {
				q.Set("phrase", tc.phrase)
			}
			if tc.limit != "" {
				q.Set("limit", tc.limit)
			}
			req.URL.RawQuery = q.Encode()

			f(recorder, req)

			require.Equal(t, tc.wantStatus, recorder.status)

			if tc.wantBodyCheck {
				require.Equal(t, "application/json", w.Header().Get("Content-Type"))

				var res SearchResponse
				err := json.NewDecoder(w.Body).Decode(&res)
				require.NoError(t, err)
				require.Equal(t, tc.wantBody, res)
			}
		})

	}
}

func TestNewPingHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	ping1 := mocks.NewMockPinger(ctrl)
	ping2 := mocks.NewMockPinger(ctrl)
	ping3 := mocks.NewMockPinger(ctrl)
	ping1.EXPECT().Ping(gomock.Any()).Return(nil)
	ping2.EXPECT().Ping(gomock.Any()).Return(nil)
	ping3.EXPECT().Ping(gomock.Any()).Return(errors.New("unavailable"))

	m := map[string]core.Pinger{
		"ping1": ping1,
		"ping2": ping2,
		"ping3": ping3,
	}

	f := NewPingHandler(logger, m)
	w := httptest.NewRecorder()
	recorder := &statusRecorder{
		ResponseWriter: w,
		status:         http.StatusOK,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	wantBody := PingResponse{
		Replies: map[string]string{
			"ping1": "ok",
			"ping2": "ok",
			"ping3": "unavailable",
		},
	}
	f(recorder, req)

	res := PingResponse{}
	require.Equal(t, http.StatusOK, recorder.status)

	err := json.NewDecoder(w.Body).Decode(&res)
	require.NoError(t, err)
	require.Equal(t, wantBody, res)
}
