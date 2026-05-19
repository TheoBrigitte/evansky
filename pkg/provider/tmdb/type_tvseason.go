package tmdb

import (
	"fmt"
	"time"

	"github.com/golusoris/goenvoy/metadata/video/tmdb"

	"github.com/TheoBrigitte/evansky/pkg/provider"
	"github.com/TheoBrigitte/evansky/pkg/util"
)

type tvSeasonResponse struct {
	*tvSeason
	multi  map[string]*tvSeason
	client *Client
}

type tvSeason struct {
	result  tmdb.SeasonDetails
	airDate time.Time
	show    provider.ResponseTV
	// Language indexed episodes cache
	episodes []provider.ResponseTVEpisode

	provider.ResponseBaseTVSeason
}

func (c *Client) newTVSeasonResponse(result tmdb.SeasonDetails, show provider.ResponseTV, req provider.Request) (*tvSeasonResponse, error) {
	m := &tvSeasonResponse{
		multi:  make(map[string]*tvSeason),
		client: c,
	}

	err := m.init(result, show, req)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (m *tvSeasonResponse) init(result tmdb.SeasonDetails, show provider.ResponseTV, req provider.Request) error {
	m.tvSeason = &tvSeason{
		result:               result,
		show:                 show,
		ResponseBaseTVSeason: provider.NewResponseBaseTVSeason(),
	}
	m.SetRequest(req)

	if result.AirDate != "" {
		// log.Debug().Msgf("parsing tv season air date: %s", result.AirDate)
		// Parse the first air date in the format "2006-01-02"
		airDate, err := time.Parse(time.DateOnly, result.AirDate)
		if err != nil {
			return err
		}
		m.airDate = airDate
	}

	languageQuery := buildLanguageQuery(req.DestinationLanguage)
	season, err := m.client.client.GetTVSeason(m.client.ctx, show.GetID(), result.SeasonNumber, languageQuery)
	if err != nil {
		return err
	}

	episodes := make([]provider.ResponseTVEpisode, 0, len(season.Episodes))
	for _, e := range season.Episodes {
		// This is a dirty way to convert an episode to tmdb.TVEpisodeDetails, as season.Episodes has no concrete type.
		var ed tmdb.EpisodeDetails
		ed.AirDate = e.AirDate
		ed.EpisodeNumber = e.EpisodeNumber
		ed.ID = e.ID
		ed.Name = e.Name
		ed.Overview = e.Overview
		ed.ProductionCode = e.ProductionCode
		ed.Runtime = e.Runtime
		ed.SeasonNumber = e.SeasonNumber
		ed.StillPath = e.StillPath
		ed.VoteAverage = e.VoteAverage
		ed.VoteCount = e.VoteCount
		episode, err := m.client.newTVEpisodeResponse(ed, m, req)
		if err != nil {
			return err
		}
		episodes = append(episodes, episode)
	}
	m.episodes = episodes
	m.multi[req.DestinationLanguage] = m.tvSeason

	return nil
}

func (r tvSeason) GetID() int {
	return int(r.result.ID)
}

func (r tvSeason) GetName() string {
	return r.result.Name
}

func (r tvSeason) GetDate() time.Time {
	return r.airDate
}

func (r tvSeason) GetProvider() string {
	return name
}

func (r tvSeason) GetPopularity() int {
	// TODO: fix this since season has no vote counts
	return util.ComputePopularity(-1, r.result.VoteAverage, 1)
}

func (r tvSeason) GetShow() provider.ResponseTV {
	return r.show
}

func (r tvSeason) GetSeasonNumber() int {
	return r.result.SeasonNumber
}

func (r tvSeason) GetEpisodes() []provider.ResponseTVEpisode {
	// slog.Debug("get episodes", "show_id", r.show.GetID(), "season_number", r.result.SeasonNumber, "episodes", len(r.episodes))
	return r.episodes
}

func (r tvSeason) GetEpisode(episodeNumber int) (provider.ResponseTVEpisode, error) {
	// slog.Debug("get episode", "show_id", r.show.GetID(), "season_number", r.result.SeasonNumber, "episode_number", episodeNumber)

	for _, e := range r.GetEpisodes() {
		if e.GetEpisodeNumber() == episodeNumber {
			return e, nil
		}
	}

	return nil, fmt.Errorf("%w for episode %d in season %d of show %d", provider.ErrNoResult, episodeNumber, r.result.SeasonNumber, r.show.GetID())
}

func (m *tvSeasonResponse) InLanguage(req provider.Request) (provider.Response, error) {
	if r, ok := m.multi[req.DestinationLanguage]; ok {
		m.tvSeason = r
	} else {
		languageQuery := buildLanguageQuery(req.DestinationLanguage)
		details, err := m.client.client.GetTVSeason(m.client.ctx, m.GetShow().GetID(), m.GetSeasonNumber(), languageQuery)
		if err != nil {
			return nil, err
		}

		err = m.init(*details, m.GetShow(), req)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}
