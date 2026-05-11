package words

import (
	"regexp"

	"github.com/kljensen/snowball/english"
)

var wordRegex = regexp.MustCompile(`[a-zA-Z0-9]+`)

func Stem(word string, stopWords bool) string {
	str := english.Stem(word, true)

	if stopWords {
		if !english.IsStopWord(str) {
			return str
		}
		return ""
	}

	return str
}

func StemSlice(phrase string, stopWords bool) []string {
	sliceWords := wordRegex.FindAllString(phrase, -1)
	var sliceAnswer []string
	set := make(map[string]struct{})

	for _, v := range sliceWords {
		str := Stem(v, stopWords)
		if _, ok := set[str]; !ok && str != "" {
			set[str] = struct{}{}
			sliceAnswer = append(sliceAnswer, str)
		}
	}
	return sliceAnswer
}
