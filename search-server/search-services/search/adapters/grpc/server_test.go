package grpc

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"yadro.com/course/search/core"
	"yadro.com/course/search/core/mocks"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	searchpb "yadro.com/course/proto/search"
)

var logger = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestServerSearch_Ping(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	searcher := mocks.NewMockSearcher(ctrl)
	s := NewServer(logger, searcher)
	res, err := s.Ping(context.Background(), new(emptypb.Empty))
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestServer_SearchMethods(t *testing.T) {
	t.Parallel()

	const (
		phrase = "I love golang so much"
		limit  = 20
	)

	type testCase struct {
		name      string
		mockSetup func(m *mocks.MockSearcher)
		call      func(s *Server) (*searchpb.SearchReply, error)
		wantErr   bool
		wantCode  codes.Code
		wantReply *searchpb.SearchReply
	}

	tests := []testCase{
		{
			name: "Search success",
			mockSetup: func(m *mocks.MockSearcher) {
				m.EXPECT().
					Search(gomock.Any(), phrase, limit).
					Return([]core.ImageInformation{
						{ID: 1, Url: "https://example.com/1.jpg"},
						{ID: 2, Url: "https://example.com/2.jpg"},
					}, nil).
					Times(1)
			},
			call: func(s *Server) (*searchpb.SearchReply, error) {
				return s.Search(context.Background(), &searchpb.SearchRequest{
					Phrase: phrase,
					Limit:  limit,
				})
			},
			wantReply: &searchpb.SearchReply{
				Images: []*searchpb.Image{
					{Id: 1, Url: "https://example.com/1.jpg"},
					{Id: 2, Url: "https://example.com/2.jpg"},
				},
			},
		},
		{
			name: "Search error",
			mockSetup: func(m *mocks.MockSearcher) {
				m.EXPECT().
					Search(gomock.Any(), phrase, limit).
					Return(nil, errors.New("db error")).
					Times(1)
			},
			call: func(s *Server) (*searchpb.SearchReply, error) {
				return s.Search(context.Background(), &searchpb.SearchRequest{
					Phrase: phrase,
					Limit:  limit,
				})
			},
			wantErr:  true,
			wantCode: codes.Internal,
		},
		{
			name: "ISearch success",
			mockSetup: func(m *mocks.MockSearcher) {
				m.EXPECT().
					ISearch(gomock.Any(), phrase, limit).
					Return([]core.ImageInformation{
						{ID: 1, Url: "https://example.com/1.jpg"},
						{ID: 2, Url: "https://example.com/2.jpg"},
					}, nil).
					Times(1)
			},
			call: func(s *Server) (*searchpb.SearchReply, error) {
				return s.ISearch(context.Background(), &searchpb.SearchRequest{
					Phrase: phrase,
					Limit:  limit,
				})
			},
			wantReply: &searchpb.SearchReply{
				Images: []*searchpb.Image{
					{Id: 1, Url: "https://example.com/1.jpg"},
					{Id: 2, Url: "https://example.com/2.jpg"},
				},
			},
		},
		{
			name: "ISearch error",
			mockSetup: func(m *mocks.MockSearcher) {
				m.EXPECT().
					ISearch(gomock.Any(), phrase, limit).
					Return(nil, errors.New("cant isearch")).
					Times(1)
			},
			call: func(s *Server) (*searchpb.SearchReply, error) {
				return s.ISearch(context.Background(), &searchpb.SearchRequest{
					Phrase: phrase,
					Limit:  limit,
				})
			},
			wantErr:  true,
			wantCode: codes.Internal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			searcher := mocks.NewMockSearcher(ctrl)
			tc.mockSetup(searcher)

			s := NewServer(logger, searcher)

			res, err := tc.call(s)

			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, res)

				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tc.wantCode, st.Code())
				require.Equal(t, "Internal error", st.Message())
				return
			}

			require.NoError(t, err)
			require.NotNil(t, res)
			require.Equal(t, tc.wantReply, res)
		})
	}
}
