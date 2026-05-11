package words

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStemSlice(t *testing.T) {
	var tests = []struct {
		words     string
		stopWords bool
		want      []string
	}{
		{
			words:     "mother",
			stopWords: false,
			want:      []string{"mother"},
		},
		{
			words:     "mother",
			stopWords: true,
			want:      []string{"mother"},
		},
		{
			words:     "at in on",
			stopWords: true,
			want:      []string(nil),
		},
		{
			words:     "at in on",
			stopWords: false,
			want:      []string{"at", "in", "on"},
		},
		{
			words:     "i will eat apple tomorrow",
			stopWords: true,
			want:      []string{"eat", "appl", "tomorrow"},
		},

		{
			words:     "i will eat apple tomorrow",
			stopWords: false,
			want:      []string{"i", "will", "eat", "appl", "tomorrow"},
		},
		{
			words:     "I had done this work</.,/ , please do not kill me",
			stopWords: true,
			want:      []string{"done", "work", "pleas", "kill"},
		},
		{
			words:     ".,. 2525 update updating server",
			stopWords: true,
			want:      []string{"2525", "updat", "server"},
		},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("Test phrase: %s   stopwords: %t", tc.words, tc.stopWords), func(t *testing.T) {
			ans := StemSlice(tc.words, tc.stopWords)
			require.Equal(t, tc.want, ans)
		})
	}
}

func TestStem(t *testing.T) {
	var tests = []struct {
		words     string
		stopWords bool
		want      string
	}{
		{
			words:     "mother",
			stopWords: false,
			want:      "mother",
		},
		{
			words:     "mother",
			stopWords: true,
			want:      "mother",
		},
		{
			words:     "at",
			stopWords: false,
			want:      "at",
		},
		{
			words:     "at",
			stopWords: true,
			want:      "",
		},
		{
			words:     "updating",
			stopWords: true,
			want:      "updat",
		},
		{
			words:     "updated",
			stopWords: true,
			want:      "updat",
		},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("Test word: %s   stopwords: %t", tc.words, tc.stopWords), func(t *testing.T) {
			ans := Stem(tc.words, tc.stopWords)
			require.Equal(t, tc.want, ans, "wrong result")
		})
	}
}
