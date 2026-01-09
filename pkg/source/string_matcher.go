package source

import (
	"strings"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/gogf/gf/v2/text/gstr"
)

func BetterMatch(a, b string, previousScore float64) (bool, float64) {
	// Normalize strings for better comparison
	a = normalizeString(a)
	b = normalizeString(b)

	// s := levenshtein(a, b)
	// isBetter := newScore < previousScore

	// newScore := jaccard(a, b)
	// newScore := overlap(a, b)
	// newScore := smithWatermanGotoh(a, b, previousScore)
	newScore := jaroWinkler(a, b)
	isBetter := newScore > previousScore

	return isBetter, newScore
}

// normalizeString normalizes a string for better matching by converting to lowercase and trimming whitespace
func normalizeString(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	return s
}

func overlap(a, b string) float64 {
	return strutil.Similarity(a, b, metrics.NewOverlapCoefficient())
}

// jaccard returns the Jaccard similarity between two strings as a float64.
func jaccard(a, b string) float64 {
	return strutil.Similarity(a, b, metrics.NewJaccard())
}

// levenshtein returns the Levenshtein distance between two strings as a float64.
func levenshtein(a, b string) float64 {
	return float64(gstr.Levenshtein(a, b, 1, 1, 1))
}

func smithWatermanGotoh(a, b string, previousScore float64) float64 {
	score := strutil.Similarity(a, b, metrics.NewSmithWatermanGotoh())
	return score
}

// jaroWinkler returns the Jaro-Winkler similarity between two strings as a float64 (0-1 range).
// Higher scores indicate better matches. This is ideal for comparing titles and names.
func jaroWinkler(a, b string) float64 {
	return strutil.Similarity(a, b, metrics.NewJaroWinkler())
}
