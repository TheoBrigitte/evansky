package source

import (
	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/gogf/gf/v2/text/gstr"
)

func betterMatch(a, b string, previousScore float64) (bool, float64) {
	//s := levenshtein(a, b)
	//isBetter := newScore < previousScore

	//newScore := jaccard(a, b)
	//newScore := overlap(a, b)
	newScore := smithWatermanGotoh(a, b, previousScore)
	isBetter := newScore > previousScore

	return isBetter, newScore
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
