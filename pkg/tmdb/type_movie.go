package tmdb

import (
	"time"

	gotmdb "github.com/cyruzin/golang-tmdb"

	"github.com/TheoBrigitte/evansky/pkg/provider"
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

func newMovie(result gotmdb.MovieResult) (*movie, error) {
	// Parse the release date in the format "2006-01-02"
	releaseDate, err := time.Parse(time.DateOnly, result.ReleaseDate)
	if err != nil {
		return nil, err
	}

	m := &movie{
		result:      result,
		releaseDate: releaseDate,
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

func (m *movieResponse) InLanguage(req provider.Request) (provider.Response, error) {
	if r, ok := m.multi[req.Language]; ok {
		m.movie = r
	} else {
		languageQuery := buildLanguageQuery(req)
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

		m.multi[req.Language] = movie
		m.movie = movie
	}

	return m, nil
}
