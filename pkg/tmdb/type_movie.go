package tmdb

import (
	"time"

	gotmdb "github.com/cyruzin/golang-tmdb"

	"github.com/TheoBrigitte/evansky/pkg/provider"
)

type movieResponse struct {
	result gotmdb.MovieResult

	releaseDate time.Time
	mediaType   provider.MediaType
}

func newMovieResponse(result gotmdb.MovieResult) (*movieResponse, error) {
	// Parse the release date in the format "2006-01-02"
	releaseDate, err := time.Parse(time.DateOnly, result.ReleaseDate)
	if err != nil {
		return nil, err
	}

	m := &movieResponse{
		result:      result,
		releaseDate: releaseDate,
		mediaType:   provider.MediaTypeMovie,
	}

	return m, nil
}

func (r movieResponse) GetID() int {
	return int(r.result.ID)
}

func (r movieResponse) GetName() string {
	return r.result.Title
}

func (r movieResponse) GetShowID() int {
	return -1
}

func (r movieResponse) GetSeasonNumber() int {
	return -1
}

func (r movieResponse) GetEpisodeNumber() int {
	return -1
}

func (r movieResponse) GetDate() time.Time {
	return r.releaseDate
}

func (r movieResponse) GetMediaType() provider.MediaType {
	return r.mediaType
}

func (r movieResponse) GetPopularity() int {
	return computePopularity(r.result.Popularity, r.result.VoteAverage, r.result.VoteCount)
}
