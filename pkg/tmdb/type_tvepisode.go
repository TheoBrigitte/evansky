package tmdb

import (
	"time"

	gotmdb "github.com/cyruzin/golang-tmdb"

	"github.com/TheoBrigitte/evansky/pkg/provider"
)

type tvEpisodeResponse struct {
	result gotmdb.TVEpisodeDetails

	showID    int
	airDate   time.Time
	mediaType provider.MediaType
}

func newTVEpisodeResponse(result gotmdb.TVEpisodeDetails, showID int) (*tvEpisodeResponse, error) {
	// Parse the first air date in the format "2006-01-02"
	airDate, err := time.Parse(time.DateOnly, result.AirDate)
	if err != nil {
		return nil, err
	}

	t := &tvEpisodeResponse{
		result:    result,
		showID:    showID,
		airDate:   airDate,
		mediaType: provider.MediaTypeTVEpisode,
	}
	return t, nil
}

func (r tvEpisodeResponse) GetID() int {
	return int(r.result.ID)
}

func (r tvEpisodeResponse) GetName() string {
	return r.result.Name
}

func (r tvEpisodeResponse) GetShowID() int {
	return r.showID
}

func (r tvEpisodeResponse) GetSeasonNumber() int {
	return r.result.SeasonNumber
}

func (r tvEpisodeResponse) GetEpisodeNumber() int {
	return r.result.EpisodeNumber
}

func (r tvEpisodeResponse) GetDate() time.Time {
	return r.airDate
}

func (r tvEpisodeResponse) GetMediaType() provider.MediaType {
	return r.mediaType
}

func (r tvEpisodeResponse) GetPopularity() int {
	return computePopularity(-1, r.result.VoteAverage, r.result.VoteCount)
}
