package tmdb

import (
	"math"
)

func computePopularity(popularity, voteAverage float64, voteCount int) int {
	if voteCount == 0 {
		return 0
	}
	return int(math.Round(float64(voteAverage) * math.Log(float64(voteCount))))
}
