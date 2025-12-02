package summary

import (
	"strings"
	"unicode"
)

// isChineseText checks if the text is primarily Chinese
func isChineseText(text string) bool {
	chineseCount := 0
	totalCount := 0
	for _, r := range text {
		if unicode.IsLetter(r) {
			totalCount++
			if unicode.Is(unicode.Han, r) {
				chineseCount++
			}
		}
	}
	if totalCount == 0 {
		return false
	}
	return float64(chineseCount)/float64(totalCount) > 0.3
}

// getTargetWordCount returns the target word count based on length setting
func getTargetWordCount(length SummaryLength) int {
	switch length {
	case Short:
		return ShortTargetWords
	case Long:
		return LongTargetWords
	default:
		return MediumTargetWords
	}
}

// countWordsOrChars counts words for English or characters for Chinese
func countWordsOrChars(text string, isChinese bool) int {
	if isChinese {
		// Count Chinese characters
		count := 0
		for _, r := range text {
			if unicode.Is(unicode.Han, r) {
				count++
			}
		}
		// Also count English words in mixed text
		englishWords := 0
		inWord := false
		for _, r := range text {
			if unicode.IsLetter(r) && !unicode.Is(unicode.Han, r) {
				if !inWord {
					englishWords++
					inWord = true
				}
			} else {
				inWord = false
			}
		}
		return count + englishWords
	}
	// Count English words
	words := strings.Fields(text)
	return len(words)
}
