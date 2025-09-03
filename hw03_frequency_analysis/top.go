package hw03frequencyanalysis

import (
	"regexp"
	"sort"
	"strings"
)

type WordFrequency struct {
	Word  string
	Count int
}

var wordRegexp = regexp.MustCompile(`[\p{L}\p{N}]+(?:-+[\p{L}\p{N}]+)*|[^-\s\p{P}]+|\p{P}{2,}`)

func Top10(text string) []string {
	if text == "" {
		return nil
	}

	freq := make(map[string]int)
	words := wordRegexp.FindAllString(text, -1)
	for _, value := range words {
		freq[strings.ToLower(value)]++
	}

	wordFrequencies := make([]WordFrequency, 0)
	for key, value := range freq {
		wordFrequencies = append(wordFrequencies, WordFrequency{Word: key, Count: value})
	}

	sort.Slice(wordFrequencies, func(i, j int) bool {
		if wordFrequencies[i].Count > wordFrequencies[j].Count {
			return true
		}
		if wordFrequencies[i].Count < wordFrequencies[j].Count {
			return false
		}
		return wordFrequencies[i].Word < wordFrequencies[j].Word
	})

	result := make([]string, 0)
	for index := range min(len(wordFrequencies), 10) {
		result = append(result, wordFrequencies[index].Word)
	}
	return result
}
