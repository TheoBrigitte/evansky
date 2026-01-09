package tmdb

import (
	"math"
	"time"

	gotmdb "github.com/cyruzin/golang-tmdb"
	"github.com/rs/zerolog/log"

	"github.com/TheoBrigitte/evansky/pkg/provider"
	"github.com/TheoBrigitte/evansky/pkg/source"
)

type movieResponse struct {
	*movie
	multi  map[string]*movie
	client *gotmdb.Client

	provider.ResponseBaseMovie
}

type movie struct {
	result      gotmdb.MovieResult
	releaseDate time.Time
}

func (c *Client) newMovieResponse(result gotmdb.MovieResult, lang string) (*movieResponse, error) {
	m, err := newMovie(result)
	if err != nil {
		return nil, err
	}

	multi := &movieResponse{
		movie:             m,
		client:            c.client,
		ResponseBaseMovie: provider.NewResponseBaseMovie(),
	}
	multi.multi = map[string]*movie{
		lang: multi.movie,
	}

	return multi, nil
}

func newMovie(result gotmdb.MovieResult) (m *movie, err error) {
	m = &movie{
		result: result,
	}

	if result.ReleaseDate != "" {
		// log.Debug().Msgf("parsing movie release date %s", result.ReleaseDate)
		// Parse the release date in the format "2006-01-02"
		m.releaseDate, err = time.Parse(time.DateOnly, result.ReleaseDate)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

func (r movie) GetID() int {
	return int(r.result.ID)
}

func (r movie) GetName() string {
	return r.result.Title
}

func (r movie) GetDate() time.Time {
	return r.releaseDate
}

func (r movie) GetPopularity() int {
	return computePopularity(r.result.Popularity, r.result.VoteAverage, r.result.VoteCount)
}

func movieByClosestYear(query string, year int, movies []gotmdb.MovieResult) (gotmdb.MovieResult, float64) {
	var bestScore float64 = -1
	var bestTitleScore float64 = 0
	var closestMatch gotmdb.MovieResult

	for index, t := range movies {
		var yearScore float64

		// Only calculate year score if year is provided
		if year > 0 {
			date, err := time.Parse(time.DateOnly, t.ReleaseDate)
			if err != nil {
				log.Warn().Err(err).Msgf("failed to parse ReleaseDate: %s", t.ReleaseDate)
				continue
			}
			yearScore = computeClosetYearScore(year, date.Year(), index)
		} else {
			// When no year is provided, use index as a small tiebreaker
			yearScore = float64(index)
		}

		// Calculate title similarity (higher is better, 0-1 range)
		_, titleScore := source.BetterMatch(query, t.Title, 0)

		// Calculate popularity score (lower is better, inverted so higher popularity = lower score)
		// Normalize popularity to 0-100 range and invert
		popularity := computePopularity(t.Popularity, t.VoteAverage, t.VoteCount)
		popularityScore := 100.0 - float64(popularity)

		// Combined score: weighted sum where title is primary, year and popularity are secondary (lower is better)
		// Title mismatch is weighted heavily (1000x) so better title matches almost always win
		// Year score and popularity score matter for breaking ties
		combinedScore := (1.0-titleScore)*1000.0 + yearScore + popularityScore

		log.Debug().Msgf("comparing movie %s tmdbid=%d date=%s yearScore=%f titleScore=%f popularity=%d popularityScore=%f combinedScore=%f",
			t.Title, t.ID, t.ReleaseDate, yearScore, titleScore, popularity, popularityScore, combinedScore)

		if bestScore == -1 || combinedScore < bestScore {
			bestScore = combinedScore
			bestTitleScore = titleScore
			closestMatch = t
		}
	}

	log.Debug().Msgf("best match: %s tmdbid=%d bestScore=%f bestTitleScore=%f",
		closestMatch.Title, closestMatch.ID, bestScore, bestTitleScore)

	return closestMatch, bestScore
}

func computeClosetYearScore(targetYear int, actualYear int, index int) float64 {
	return (math.Abs(float64(targetYear-actualYear)) + 1) * (math.Exp(float64(index + 2)))
}

func (m *movieResponse) InLanguage(req provider.Request) (provider.Response, error) {
	if r, ok := m.multi[req.DestinationLanguage]; ok {
		m.movie = r
	} else {
		languageQuery := buildLanguageQuery(req.DestinationLanguage)
		details, err := m.client.GetMovieDetails(m.GetID(), languageQuery)
		if err != nil {
			return nil, err
		}

		// TODO: fetch the movie details in newMovie, so we can also store the full details here
		result := gotmdb.MovieResult{
			ID:               details.ID,
			Title:            details.Title,
			OriginalTitle:    details.OriginalTitle,
			OriginalLanguage: details.OriginalLanguage,
			Overview:         details.Overview,
			ReleaseDate:      details.ReleaseDate,
			PosterPath:       details.PosterPath,
			BackdropPath:     details.BackdropPath,
			Popularity:       details.Popularity,
			Adult:            details.Adult,
			Video:            details.Video,
			VoteMetrics:      details.VoteMetrics,
		}

		movie, err := newMovie(result)
		if err != nil {
			return nil, err
		}

		m.multi[req.DestinationLanguage] = movie
		m.movie = movie
	}

	return m, nil
}
