package tmdb

import (
	"fmt"
	"log/slog"
	"time"

	gotmdb "github.com/cyruzin/golang-tmdb"

	"github.com/TheoBrigitte/evansky/pkg/provider"
)

type tvSeasonResponse struct {
	*tvSeason
	multi  map[string]*tvSeason
	client *Client
	provider.ResponseBaseTVSeason
}

type tvSeason struct {
	result  gotmdb.TVSeasonDetails
	airDate time.Time
	show    provider.ResponseTV
	// Language indexed episodes cache
	episodes []provider.ResponseTVEpisode
}

func (c *Client) newTVSeasonResponse(result gotmdb.TVSeasonDetails, show provider.ResponseTV, req provider.Request) (*tvSeasonResponse, error) {
	m := &tvSeasonResponse{
		multi:                make(map[string]*tvSeason),
		client:               c,
		ResponseBaseTVSeason: provider.NewResponseBaseTVSeason(),
	}

	err := m.init(result, show, req)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (m *tvSeasonResponse) init(result gotmdb.TVSeasonDetails, show provider.ResponseTV, req provider.Request) error {
	m.tvSeason = &tvSeason{
		result: result,
		show:   show,
	}

	// Parse the first air date in the format "2006-01-02"
	airDate, err := time.Parse(time.DateOnly, result.AirDate)
	if err != nil {
		return err
	}
	m.airDate = airDate

	languageQuery := buildLanguageQuery(req)
	season, err := m.client.client.GetTVSeasonDetails(show.GetID(), result.SeasonNumber, languageQuery)
	if err != nil {
		return err
	}

	episodes := make([]provider.ResponseTVEpisode, 0, len(season.Episodes))
	for _, e := range season.Episodes {
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
		episode, err := m.client.newTVEpisodeResponse(ed, m, req.Language)
		if err != nil {
			return err
		}
		episodes = append(episodes, episode)
	}
	m.episodes = episodes
	m.multi[req.Language] = m.tvSeason

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

func (r tvSeason) GetPopularity() int {
	// TODO: fix this since season has no vote counts
	return computePopularity(-1, r.result.VoteAverage, 1)
}

func (r tvSeason) GetShow() provider.ResponseTV {
	return r.show
}

func (r tvSeason) GetSeasonNumber() int {
	return r.result.SeasonNumber
}

func (r tvSeason) GetEpisodes() []provider.ResponseTVEpisode {
	slog.Debug("get episodes", "show_id", r.show.GetID(), "season_number", r.result.SeasonNumber, "episodes", len(r.episodes))
	return r.episodes
}

func (r tvSeason) GetEpisode(episodeNumber int) (provider.ResponseTVEpisode, error) {
	slog.Debug("get episode", "show_id", r.show.GetID(), "season_number", r.result.SeasonNumber, "episode_number", episodeNumber)

	for _, e := range r.GetEpisodes() {
		if e.GetEpisodeNumber() == episodeNumber {
			return e, nil
		}
	}

	return nil, fmt.Errorf("episode %d not found in season %d of show %d", episodeNumber, r.result.SeasonNumber, r.show.GetID())
}

func (m *tvSeasonResponse) InLanguage(req provider.Request) (provider.Response, error) {
	if r, ok := m.multi[req.Language]; ok {
		m.tvSeason = r
	} else {
		languageQuery := buildLanguageQuery(req)
		details, err := m.client.client.GetTVSeasonDetails(m.GetShow().GetID(), m.GetSeasonNumber(), languageQuery)
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
