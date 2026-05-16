package tmdb

import (
	"time"

	"github.com/golusoris/goenvoy/metadata/video/tmdb"

	"github.com/TheoBrigitte/evansky/pkg/provider"
	"github.com/TheoBrigitte/evansky/pkg/util"
)

type movieResponse struct {
	*movie
	multi  map[string]*movie
	client *Client
}

type movie struct {
	result      tmdb.MovieResult
	releaseDate time.Time

	provider.ResponseBaseMovie
}

func (c *Client) newMovieResponse(result tmdb.MovieResult, req provider.Request) (*movieResponse, error) {
	m, err := newMovie(result, req)
	if err != nil {
		return nil, err
	}

	r := &movieResponse{
		movie:  m,
		client: c,
	}
	r.multi = map[string]*movie{
		req.QueryLanguage: r.movie,
	}

	return r, nil
}

func newMovie(result tmdb.MovieResult, req provider.Request) (m *movie, err error) {
	m = &movie{
		result:            result,
		ResponseBaseMovie: provider.NewResponseBaseMovie(),
	}
	m.SetRequest(req)

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
	return util.ComputePopularity(r.result.Popularity, r.result.VoteAverage, r.result.VoteCount)
}

func (r movie) GetProvider() string {
	return name
}

func (m *movieResponse) InLanguage(req provider.Request) (provider.Response, error) {
	if r, ok := m.multi[req.DestinationLanguage]; ok {
		m.movie = r
	} else {
		languageQuery := buildLanguageQuery(req.DestinationLanguage)
		details, err := m.client.client.GetMovie(m.client.ctx, m.GetID(), languageQuery)
		if err != nil {
			return nil, err
		}

		// TODO: fetch the movie details in newMovie, so we can also store the full details here
		result := tmdb.MovieResult{
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
			VoteAverage:      details.VoteAverage,
			VoteCount:        details.VoteCount,
		}

		movie, err := newMovie(result, req)
		if err != nil {
			return nil, err
		}

		m.multi[req.DestinationLanguage] = movie
		m.movie = movie
	}

	return m, nil
}
