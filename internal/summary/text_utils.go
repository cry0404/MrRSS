package summary

import (
	"regexp"
	"strings"
	"sync"
	"unicode"

	"github.com/go-ego/gse"
)

// Global segmenter instance with lazy initialization
var (
	segmenter     gse.Segmenter
	segmenterOnce sync.Once
)

// getSegmenter returns the global segmenter, initializing it if necessary
func getSegmenter() *gse.Segmenter {
	segmenterOnce.Do(func() {
		// Load default dictionary for Chinese segmentation
		segmenter.LoadDict()
	})
	return &segmenter
}

// cleanText removes HTML tags and normalizes whitespace
func cleanText(text string) string {
	// Remove HTML tags
	htmlRegex := regexp.MustCompile(`<[^>]*>`)
	text = htmlRegex.ReplaceAllString(text, " ")

	// Decode common HTML entities
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")

	// Normalize whitespace
	spaceRegex := regexp.MustCompile(`\s+`)
	text = spaceRegex.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

// splitSentences splits text into sentences
func splitSentences(text string) []string {
	// Simple sentence splitting with common abbreviations handling
	// Split on sentence-ending punctuation followed by space (or end of text)
	sentenceRegex := regexp.MustCompile(`([.!?。！？]+)(\s+|$)`)

	// Split the text
	parts := sentenceRegex.Split(text, -1)

	// Get the delimiters
	matches := sentenceRegex.FindAllStringSubmatch(text, -1)

	var sentences []string
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Add back the punctuation if available
		if i < len(matches) && len(matches[i]) > 1 {
			part += matches[i][1]
		}

		// Filter out very short sentences (likely fragments)
		// Use a lower threshold to support various languages
		if len(part) > 10 {
			sentences = append(sentences, part)
		}
	}

	return sentences
}

// tokenize splits text into lowercase tokens, removing stopwords
// Uses gse for Chinese word segmentation for better accuracy
func tokenize(text string) []string {
	// Convert to lowercase for English
	text = strings.ToLower(text)

	// Check if text contains Chinese characters
	hasChinese := false
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			hasChinese = true
			break
		}
	}

	var tokens []string

	if hasChinese {
		// Use gse for Chinese text segmentation
		seg := getSegmenter()
		segments := seg.Cut(text, true) // true = search mode for better recall

		for _, word := range segments {
			word = strings.TrimSpace(word)
			// Skip empty strings, stopwords, and very short words
			if len(word) > 0 && !isStopWord(word) {
				// For Chinese, single characters can be meaningful
				// For English, require at least 2 characters
				isChinese := false
				for _, r := range word {
					if unicode.Is(unicode.Han, r) {
						isChinese = true
						break
					}
				}
				if isChinese || len(word) > 2 {
					tokens = append(tokens, word)
				}
			}
		}
	} else {
		// Use simple tokenization for non-Chinese text
		var currentWord strings.Builder

		for _, r := range text {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				currentWord.WriteRune(r)
			} else {
				if currentWord.Len() > 0 {
					word := currentWord.String()
					// Skip stopwords and very short words
					if len(word) > 2 && !isStopWord(word) {
						tokens = append(tokens, word)
					}
					currentWord.Reset()
				}
			}
		}

		// Don't forget the last word
		if currentWord.Len() > 0 {
			word := currentWord.String()
			if len(word) > 2 && !isStopWord(word) {
				tokens = append(tokens, word)
			}
		}
	}

	return tokens
}

// isStopWord checks if a word is a common stopword (English and Chinese)
func isStopWord(word string) bool {
	stopWords := map[string]bool{
		// English stopwords
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "from": true, "as": true, "is": true, "was": true,
		"are": true, "were": true, "been": true, "be": true, "have": true, "has": true,
		"had": true, "do": true, "does": true, "did": true, "will": true, "would": true,
		"could": true, "should": true, "may": true, "might": true, "must": true,
		"shall": true, "can": true, "this": true, "that": true, "these": true,
		"those": true, "it": true, "its": true, "they": true, "them": true,
		"their": true, "what": true, "which": true, "who": true, "whom": true,
		"whose": true, "where": true, "when": true, "why": true, "how": true,
		"all": true, "each": true, "every": true, "both": true, "few": true,
		"more": true, "most": true, "other": true, "some": true, "such": true,
		"than": true, "too": true, "very": true, "just": true, "only": true,
		"own": true, "same": true, "so": true, "not": true, "also": true,
		"into": true, "about": true, "your": true, "you": true, "our": true,
		"his": true, "her": true, "my": true, "we": true, "he": true, "she": true,
		"over": true, "out": true, "up": true, "down": true, "then": true, "now": true,
		// Extended Chinese stopwords for better accuracy
		"的": true, "了": true, "和": true, "是": true, "在": true, "有": true,
		"这": true, "个": true, "我": true, "不": true, "人": true, "都": true,
		"一": true, "他": true, "就": true, "们": true, "上": true, "也": true,
		"你": true, "说": true, "着": true, "对": true, "为": true, "与": true,
		"而": true, "等": true, "被": true, "把": true, "让": true, "给": true,
		"向": true, "从": true, "到": true, "之": true, "于": true, "或": true,
		"因": true, "但": true, "却": true, "即": true, "若": true, "虽": true,
		"所": true, "以": true, "如": true, "则": true, "其": true, "它": true,
		"她": true, "这个": true, "那个": true, "什么": true, "怎么": true,
		"为什么": true, "哪个": true, "哪些": true, "这些": true, "那些": true,
		"可以": true, "能够": true, "已经": true, "正在": true, "将要": true,
		"可能": true, "应该": true, "必须": true, "需要": true, "没有": true,
		"因为": true, "所以": true, "但是": true, "而且": true, "或者": true,
		"如果": true, "虽然": true, "然": true, "此": true, "彼": true,
		"自己": true, "我们": true, "你们": true, "他们": true, "它们": true,
		"这里": true, "那里": true, "哪里": true, "任何": true, "某些": true,
		"每个": true, "很": true, "非常": true, "十分": true, "比较": true,
		"更": true, "最": true, "太": true, "又": true, "再": true, "还": true,
	}
	return stopWords[word]
}
