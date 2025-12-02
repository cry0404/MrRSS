package summary

import (
	"math"
)

// calculateTFIDF computes TF-IDF scores for each sentence
func calculateTFIDF(sentences []string) []float64 {
	// Build document frequency map
	docFreq := make(map[string]int)
	allTerms := make([]map[string]int, len(sentences))

	for i, sentence := range sentences {
		terms := tokenize(sentence)
		termFreq := make(map[string]int)
		seenTerms := make(map[string]bool)

		for _, term := range terms {
			termFreq[term]++
			if !seenTerms[term] {
				docFreq[term]++
				seenTerms[term] = true
			}
		}
		allTerms[i] = termFreq
	}

	numDocs := float64(len(sentences))
	scores := make([]float64, len(sentences))

	for i, termFreq := range allTerms {
		var score float64
		totalTerms := 0
		for _, count := range termFreq {
			totalTerms += count
		}

		for term, count := range termFreq {
			// TF: normalized term frequency
			tf := float64(count) / float64(totalTerms)

			// IDF: inverse document frequency
			idf := math.Log(numDocs / float64(docFreq[term]))

			score += tf * idf
		}
		scores[i] = score
	}

	// Normalize scores
	maxScore := 0.0
	for _, score := range scores {
		if score > maxScore {
			maxScore = score
		}
	}
	if maxScore > 0 {
		for i := range scores {
			scores[i] /= maxScore
		}
	}

	return scores
}

// calculateTextRank computes TextRank scores using sentence similarity
func calculateTextRank(sentences []string) []float64 {
	n := len(sentences)
	if n == 0 {
		return []float64{}
	}

	// Build similarity matrix
	similarity := make([][]float64, n)
	for i := range similarity {
		similarity[i] = make([]float64, n)
	}

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			sim := sentenceSimilarity(sentences[i], sentences[j])
			similarity[i][j] = sim
			similarity[j][i] = sim
		}
	}

	// Initialize scores
	scores := make([]float64, n)
	for i := range scores {
		scores[i] = 1.0
	}

	// Damping factor
	d := 0.85
	iterations := 30

	// Iterate to convergence
	for iter := 0; iter < iterations; iter++ {
		newScores := make([]float64, n)
		for i := 0; i < n; i++ {
			sum := 0.0
			for j := 0; j < n; j++ {
				if i != j {
					// Calculate sum of similarities for sentence j
					sumSim := 0.0
					for k := 0; k < n; k++ {
						if k != j {
							sumSim += similarity[j][k]
						}
					}
					if sumSim > 0 {
						sum += similarity[i][j] / sumSim * scores[j]
					}
				}
			}
			newScores[i] = (1 - d) + d*sum
		}
		scores = newScores
	}

	// Normalize scores
	maxScore := 0.0
	for _, score := range scores {
		if score > maxScore {
			maxScore = score
		}
	}
	if maxScore > 0 {
		for i := range scores {
			scores[i] /= maxScore
		}
	}

	return scores
}

// sentenceSimilarity calculates similarity between two sentences using word overlap
func sentenceSimilarity(s1, s2 string) float64 {
	words1 := tokenize(s1)
	words2 := tokenize(s2)

	if len(words1) == 0 || len(words2) == 0 {
		return 0
	}

	// Create word sets
	set1 := make(map[string]bool)
	for _, w := range words1 {
		set1[w] = true
	}

	set2 := make(map[string]bool)
	for _, w := range words2 {
		set2[w] = true
	}

	// Guard against empty sets to avoid math.Log(0) which returns -Inf
	if len(set1) == 0 || len(set2) == 0 {
		return 0
	}

	// Count common words
	common := 0
	for w := range set1 {
		if set2[w] {
			common++
		}
	}

	// Normalized overlap
	denom := math.Log(float64(len(set1))) + math.Log(float64(len(set2)))
	if denom == 0 {
		return 0
	}

	return float64(common) / denom
}
