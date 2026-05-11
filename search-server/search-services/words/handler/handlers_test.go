package handler

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
	wordspb "yadro.com/course/proto/words"
)

func TestServer_Ping(t *testing.T) {
	s := NewServer(slog.Default())

	ans, err := s.Ping(context.Background(), &emptypb.Empty{})
	require.NoError(t, err)
	require.NotNil(t, ans)
}

func TestServer_Norm_OK(t *testing.T) {
	s := NewServer(slog.Default())

	resp, err := s.Norm(context.Background(), &wordspb.WordsRequest{
		Phrase: "I had done this work, please do not kill me",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, []string{"done", "work", "pleas", "kill"}, resp.Words)
}

func TestServer_Norm_BigMassge(t *testing.T) {
	s := NewServer(slog.Default())

	phrase := strings.Repeat("a", maxMessageSize+1)
	resp, err := s.Norm(context.Background(), &wordspb.WordsRequest{
		Phrase: phrase,
	})
	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "too much")
}
