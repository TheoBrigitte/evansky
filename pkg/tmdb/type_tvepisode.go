package tmdb

import (
	"time"

	gotmdb "github.com/cyruzin/golang-tmdb"

	"github.com/TheoBrigitte/evansky/pkg/provider"
)

type tvEpisodeResponse struct {
	provider.ResponseBaseTVEpisode
	result gotmdb.TVEpisodeDetails

	airDate time.Time
	season  provider.ResponseTVSeason
	client  *gotmdb.Client
}

func (c *Client) newTVEpisodeResponse(result gotmdb.TVEpisodeDetails, season provider.ResponseTVSeason) (*tvEpisodeResponse, error) {
	// Parse the first air date in the format "2006-01-02"
	airDate, err := time.Parse(time.DateOnly, result.AirDate)
	if err != nil {
		return nil, err
	}

	t := &tvEpisodeResponse{
		ResponseBaseTVEpisode: provider.NewResponseBaseTVEpisode(),
		result:                result,
		airDate:               airDate,
		season:                season,
		client:                c.client,
	}
	return t, nil
}

func (r tvEpisodeResponse) GetID() int {
	return int(r.result.ID)
}

func (r tvEpisodeResponse) GetName() string {
	return r.result.Name
}

func (r tvEpisodeResponse) GetDate() time.Time {
	return r.airDate
}

func (r tvEpisodeResponse) GetPopularity() int {
	return computePopularity(-1, r.result.VoteAverage, r.result.VoteCount)
}

func (r tvEpisodeResponse) GetEpisodeNumber() int {
	return r.result.EpisodeNumber
}

func (r tvEpisodeResponse) GetSeason() provider.ResponseTVSeason {
	return r.season
}
