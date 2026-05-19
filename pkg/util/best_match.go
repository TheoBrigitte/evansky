package util

import (
	"math"

	"github.com/rs/zerolog/log"

	"github.com/TheoBrigitte/evansky/pkg/provider"
	"github.com/TheoBrigitte/evansky/pkg/source"
)

// BestMatch compares a list of elements to a search request and returns the closest match based on title similarity, year proximity, and popularity.
// TODO: add error in return to handle cases where no match is found
func BestMatch[E any, R provider.Response](req provider.Request, elements []E, newE func(E, provider.Request) (R, error)) (R, float64) {
	var bestScore float64 = -1
	var bestTitleScore float64 = 0
	var closestMatch R

	for index, t := range elements {
		var yearScore float64

		e, err := newE(t, req)
		if err != nil {
			log.Warn().Err(err).Msgf("failed to convert element to response: %v", t)
		}

		// Only calculate year score if year is provided
		if req.Year > 0 {
			yearScore = computeClosetYearScore(req.Year, e.GetDate().Year(), index)
		} else {
			// When no year is provided, use index as a small tiebreaker
			yearScore = float64(index)
		}

		// Calculate title similarity (higher is better, 0-1 range)
		_, titleScore := source.BetterMatch(req.Query, e.GetName(), 0)

		// Calculate popularity score (lower is better, inverted so higher popularity = lower score)
		// Normalize popularity to 0-100 range and invert
		popularity := e.GetPopularity()
		popularityScore := 100.0 - float64(popularity)

		// Combined score: weighted sum where title is primary, year and popularity are secondary (lower is better)
		// Title mismatch is weighted heavily (1000x) so better title matches almost always win
		// Year score and popularity score matter for breaking ties
		combinedScore := (1.0-titleScore)*1000.0 + yearScore + popularityScore

		log.Debug().Msgf("comparing %T title=%s provider=%s providerId=%d date=%s yearScore=%f titleScore=%f popularity=%d popularityScore=%f combinedScore=%f",
			t, e.GetName(), e.GetProvider(), e.GetID(), e.GetDate(), yearScore, titleScore, popularity, popularityScore, combinedScore)

		if bestScore == -1 || combinedScore < bestScore {
			bestScore = combinedScore
			bestTitleScore = titleScore
			closestMatch = e
		}
	}

	log.Debug().Msgf("best match: title=%s provider=%s providerId=%d bestScore=%f bestTitleScore=%f",
		closestMatch.GetName(), closestMatch.GetProvider(), closestMatch.GetID(), bestScore, bestTitleScore)

	return closestMatch, bestScore
}

func computeClosetYearScore(targetYear int, actualYear int, index int) float64 {
	return (math.Abs(float64(targetYear-actualYear)) + 1) * (math.Exp(float64(index + 2)))
}
