package tmdb

import (
	"fmt"
	"log/slog"
	"time"

	gotmdb "github.com/cyruzin/golang-tmdb"

	"github.com/TheoBrigitte/evansky/pkg/provider"
)

type tvSeasonResponse struct {
	provider.ResponseBaseTVSeason
	result gotmdb.TVSeasonDetails

	airDate time.Time
	show    provider.ResponseTV
	client  *Client
}

func (c *Client) newTVSeasonResponse(result gotmdb.TVSeasonDetails, show provider.ResponseTV) (*tvSeasonResponse, error) {
	// Parse the first air date in the format "2006-01-02"
	airDate, err := time.Parse(time.DateOnly, result.AirDate)
	if err != nil {
		return nil, err
	}

	t := &tvSeasonResponse{
		ResponseBaseTVSeason: provider.NewResponseBaseTVSeason(),
		result:               result,
		show:                 show,
		airDate:              airDate,
		client:               c,
	}
	return t, nil
}

func (r tvSeasonResponse) GetID() int {
	return int(r.result.ID)
}

func (r tvSeasonResponse) GetName() string {
	return r.result.Name
}

func (r tvSeasonResponse) GetDate() time.Time {
	return r.airDate
}

func (r tvSeasonResponse) GetPopularity() int {
	// TODO: fix this since season has no vote counts
	return computePopularity(-1, r.result.VoteAverage, 1)
}

func (r tvSeasonResponse) GetShow() provider.ResponseTV {
	return r.show
}

func (r tvSeasonResponse) GetSeasonNumber() int {
	return r.result.SeasonNumber
}

func (r tvSeasonResponse) GetEpisodes() ([]provider.ResponseTVEpisode, error) {
	slog.Debug("get episodes", "show_id", r.show.GetID(), "season_number", r.result.SeasonNumber)
	episodes := make([]provider.ResponseTVEpisode, 0, len(r.result.Episodes))
	for _, e := range r.result.Episodes {
		// This is a dirty way to convert an episode to gotmdb.TVEpisodeDetails, as season.Episodes has no concrete type.
		var ed gotmdb.TVEpisodeDetails
		ed.AirDate = e.AirDate
		ed.EpisodeNumber = e.EpisodeNumber
		ed.ID = e.ID
		ed.Name = e.Name
		ed.Overview = e.Overview
		ed.ProductionCode = e.ProductionCode
		ed.Runtime = e.Runtime
		ed.SeasonNumber = e.SeasonNumber
		ed.StillPath = e.StillPath
		ed.VoteMetrics = e.VoteMetrics
		ed.Crew = e.Crew
		ed.GuestStars = e.GuestStars
		episode, err := r.client.newTVEpisodeResponse(ed, r)
		if err != nil {
			return nil, err
		}
		episodes = append(episodes, episode)
	}

	if len(episodes) == 0 {
		return nil, fmt.Errorf("no episode found in season %d of show %d", r.result.SeasonNumber, r.show.GetID())
	}

	return episodes, nil
}

func (r tvSeasonResponse) GetEpisode(episodeNumber int) (provider.ResponseTVEpisode, error) {
	slog.Debug("get episode", "show_id", r.show.GetID(), "season_number", r.result.SeasonNumber, "episode_number", episodeNumber)
	for _, e := range r.result.Episodes {
		if e.EpisodeNumber == episodeNumber {
			// This is a dirty way to convert an episode to gotmdb.TVEpisodeDetails, as season.Episodes has no concrete type.
			var ed gotmdb.TVEpisodeDetails
			ed.AirDate = e.AirDate
			ed.EpisodeNumber = e.EpisodeNumber
			ed.ID = e.ID
			ed.Name = e.Name
			ed.Overview = e.Overview
			ed.ProductionCode = e.ProductionCode
			ed.Runtime = e.Runtime
			ed.SeasonNumber = e.SeasonNumber
			ed.StillPath = e.StillPath
			ed.VoteMetrics = e.VoteMetrics
			ed.Crew = e.Crew
			ed.GuestStars = e.GuestStars
			return r.client.newTVEpisodeResponse(ed, r)
		}
	}

	return nil, fmt.Errorf("episode %d not found in season %d of show %d", episodeNumber, r.result.SeasonNumber, r.show.GetID())
}
