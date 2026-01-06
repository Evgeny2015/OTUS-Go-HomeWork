package hw03frequencyanalysis

import (
	"sort"
	"strings"
)

type WordCount struct {
	Word  string
	Count int
}

const TopWordCount = 10

func Top10(str string) []string {
	// get words from string
	words := strings.Fields(str)
	if len(words) == 0 {
		return nil
	}

	// count words
	dict := map[string]int{}
	for _, s := range words {
		s = strings.ToLower(strings.Trim(s, "',!."))
		dict[s]++
	}

	// remove invalid string
	delete(dict, "-")

	// convert to slice
	wordSlice := make([]WordCount, 0, len(dict))
	for key, value := range dict {
		wordSlice = append(wordSlice, WordCount{Word: key, Count: value})
	}

	// sort words
	sort.Slice(wordSlice, func(i, j int) bool {
		if wordSlice[i].Count == wordSlice[j].Count {
			return wordSlice[i].Word < wordSlice[j].Word
		}

		return wordSlice[i].Count > wordSlice[j].Count
	})

	// get top 10 words
	res := make([]string, 0, TopWordCount)
	for i := 0; i < min(TopWordCount, len(wordSlice)-1); i++ {
		res = append(res, wordSlice[i].Word)
	}

	return res
}
