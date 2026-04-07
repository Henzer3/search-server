package words

import (
	"regexp"

	"github.com/kljensen/snowball/english"
)

var wordRegex = regexp.MustCompile(`[a-zA-Z0-9]+`)

func Stem(word string, stopWords bool) (string, error) {
	str := english.Stem(word, true)

	if stopWords {
		if !english.IsStopWord(str) {
			return str, nil
		}
		return "", nil
	}

	return str, nil
}

func StemSlice(phrase string, stopWords bool) ([]string, error) {
	sliseWords := wordRegex.FindAllString(phrase, -1)
	var sliseAnswer []string
	set := make(map[string]struct{})

	for _, v := range sliseWords {
		str, err := Stem(v, true)
		if err != nil {
			return nil, err
		}
		if _, ok := set[str]; !ok && str != "" {
			set[str] = struct{}{}
			sliseAnswer = append(sliseAnswer, str)
		}
	}
	return sliseAnswer, nil
}
