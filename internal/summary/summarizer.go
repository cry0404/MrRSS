// Package summary provides text summarization using local algorithms.
// It implements TF-IDF and TextRank-based sentence scoring for extractive summarization.
package summary

import (
	"sort"
	"strings"
)

// Summarizer provides text summarization capabilities
type Summarizer struct{}

// NewSummarizer creates a new Summarizer instance
func NewSummarizer() *Summarizer {
	return &Summarizer{}
}

// Summarize generates a summary of the given text using combined TF-IDF and TextRank scoring
func (s *Summarizer) Summarize(text string, length SummaryLength) SummaryResult {
	// Clean the text
	cleanedText := cleanText(text)

	// Check if text is too short
	if len(cleanedText) < MinContentLength {
		return SummaryResult{
			Summary:    cleanedText,
			IsTooShort: true,
		}
	}

	// Split into sentences
	sentences := splitSentences(cleanedText)

	// Check if we have enough sentences
	if len(sentences) < MinSentenceCount {
		return SummaryResult{
			Summary:       cleanedText,
			SentenceCount: len(sentences),
			IsTooShort:    true,
		}
	}

	// Check if text is primarily Chinese
	isChinese := isChineseText(cleanedText)

	// Get target word/character count based on length setting
	targetCount := getTargetWordCount(length)

	// Score sentences using combined TF-IDF and TextRank
	scoredSentences := s.scoreSentences(sentences)

	// Sort by score (descending)
	sort.Slice(scoredSentences, func(i, j int) bool {
		return scoredSentences[i].score > scoredSentences[j].score
	})

	// Select sentences until we reach the target word/character count
	var selectedSentences []scoredSentence
	currentCount := 0

	for _, sent := range scoredSentences {
		sentCount := countWordsOrChars(sent.text, isChinese)
		if currentCount+sentCount <= targetCount || len(selectedSentences) == 0 {
			selectedSentences = append(selectedSentences, sent)
			currentCount += sentCount
		}
		// Stop if we've reached the target
		if currentCount >= targetCount {
			break
		}
	}

	// Sort by original position to maintain narrative flow
	sort.Slice(selectedSentences, func(i, j int) bool {
		return selectedSentences[i].position < selectedSentences[j].position
	})

	// Build summary
	var summaryParts []string
	for _, sent := range selectedSentences {
		summaryParts = append(summaryParts, sent.text)
	}

	return SummaryResult{
		Summary:       strings.Join(summaryParts, " "),
		SentenceCount: len(selectedSentences),
		IsTooShort:    false,
	}
}

// scoreSentences calculates scores for each sentence using combined TF-IDF and TextRank
func (s *Summarizer) scoreSentences(sentences []string) []scoredSentence {
	// Calculate TF-IDF scores
	tfidfScores := calculateTFIDF(sentences)

	// Calculate TextRank scores
	textRankScores := calculateTextRank(sentences)

	// Calculate average sentence length for penalty calculation
	totalLen := 0
	for _, sent := range sentences {
		totalLen += len(sent)
	}
	avgLen := float64(totalLen) / float64(len(sentences))

	// Combine scores
	result := make([]scoredSentence, len(sentences))
	for i, sentence := range sentences {
		// Weight TF-IDF at 0.5 and TextRank at 0.5
		combinedScore := 0.5*tfidfScores[i] + 0.5*textRankScores[i]

		// Boost first sentence slightly (often contains key info)
		if i == 0 {
			combinedScore *= 1.15
		}

		// Penalize very long sentences (more than 2x average length)
		// This prevents selecting overly verbose sentences
		sentLen := float64(len(sentence))
		if sentLen > avgLen*2 {
			penalty := avgLen * 2 / sentLen
			combinedScore *= penalty
		}

		// Slight penalty for very short sentences (less than 0.3x average)
		// They often lack sufficient information
		if sentLen < avgLen*0.3 {
			combinedScore *= 0.8
		}

		result[i] = scoredSentence{
			text:     sentence,
			score:    combinedScore,
			position: i,
		}
	}

	return result
}
