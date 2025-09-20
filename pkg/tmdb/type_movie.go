package tmdb

import (
	"time"

	gotmdb "github.com/cyruzin/golang-tmdb"

	"github.com/TheoBrigitte/evansky/pkg/provider"
)

type movieResponse struct {
	provider.ResponseBaseMovie

	result      gotmdb.MovieResult
	releaseDate time.Time
	client      *gotmdb.Client
}

func (c *Client) newMovieResponse(result gotmdb.MovieResult) (*movieResponse, error) {
	// Parse the release date in the format "2006-01-02"
	releaseDate, err := time.Parse(time.DateOnly, result.ReleaseDate)
	if err != nil {
		return nil, err
	}

	m := &movieResponse{
		ResponseBaseMovie: provider.NewResponseBaseMovie(),
		result:            result,
		releaseDate:       releaseDate,
		client:            c.client,
	}

	return m, nil
}

func (r movieResponse) GetID() int {
	return int(r.result.ID)
}

func (r movieResponse) GetName() string {
	return r.result.Title
}

func (r movieResponse) GetDate() time.Time {
	return r.releaseDate
}

func (r movieResponse) GetPopularity() int {
	return computePopularity(r.result.Popularity, r.result.VoteAverage, r.result.VoteCount)
}
