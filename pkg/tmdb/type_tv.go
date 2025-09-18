package tmdb

import (
	"time"

	gotmdb "github.com/cyruzin/golang-tmdb"

	"github.com/TheoBrigitte/evansky/pkg/provider"
)

type tvResponse struct {
	result gotmdb.TVShowResult

	firstAirDate time.Time
	mediaType    provider.MediaType
}

func newTVResponse(result gotmdb.TVShowResult) (*tvResponse, error) {
	// Parse the first air date in the format "2006-01-02"
	firstAirDate, err := time.Parse(time.DateOnly, result.FirstAirDate)
	if err != nil {
		return nil, err
	}

	t := &tvResponse{
		result:       result,
		firstAirDate: firstAirDate,
		mediaType:    provider.MediaTypeTV,
	}
	return t, nil
}

func (r tvResponse) GetID() int {
	return int(r.result.ID)
}

func (r tvResponse) GetName() string {
	return r.result.Name
}

func (r tvResponse) GetSeasonNumber() int {
	return -1
}

func (r tvResponse) GetEpisodeNumber() int {
	return -1
}

func (r tvResponse) GetDate() time.Time {
	return r.firstAirDate
}

func (r tvResponse) GetMediaType() provider.MediaType {
	return r.mediaType
}

func (r tvResponse) GetPopularity() int {
	return computePopularity(r.result.Popularity, r.result.VoteAverage, r.result.VoteCount)
}
