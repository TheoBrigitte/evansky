package tmdb

import (
	"time"

	gotmdb "github.com/cyruzin/golang-tmdb"

	"github.com/TheoBrigitte/evansky/pkg/provider"
)

type tvSeasonResponse struct {
	result gotmdb.TVSeasonDetails

	showID    int
	airDate   time.Time
	mediaType provider.MediaType
}

func newTVSeasonResponse(result gotmdb.TVSeasonDetails, showID int) (*tvSeasonResponse, error) {
	// Parse the first air date in the format "2006-01-02"
	airDate, err := time.Parse(time.DateOnly, result.AirDate)
	if err != nil {
		return nil, err
	}

	t := &tvSeasonResponse{
		result:    result,
		showID:    showID,
		airDate:   airDate,
		mediaType: provider.MediaTypeTVSeason,
	}
	return t, nil
}

func (r tvSeasonResponse) GetID() int {
	return int(r.result.ID)
}

func (r tvSeasonResponse) GetName() string {
	return r.result.Name
}

func (r tvSeasonResponse) GetShowID() int {
	return r.showID
}

func (r tvSeasonResponse) GetSeasonNumber() int {
	return r.result.SeasonNumber
}

func (r tvSeasonResponse) GetEpisodeNumber() int {
	return -1
}

func (r tvSeasonResponse) GetDate() time.Time {
	return r.airDate
}

func (r tvSeasonResponse) GetMediaType() provider.MediaType {
	return r.mediaType
}

func (r tvSeasonResponse) GetPopularity() int {
	// TODO: fix this since season has no vote counts
	return computePopularity(-1, r.result.VoteAverage, 1)
}
