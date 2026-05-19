package tmdb

import (
	"context"
	"time"

	"github.com/golusoris/goenvoy/metadata/video/tmdb"

	"github.com/TheoBrigitte/evansky/pkg/provider"
	"github.com/TheoBrigitte/evansky/pkg/util"
)

type tvEpisodeResponse struct {
	*tvEpisode
	multi  map[string]*tvEpisode
	client *tmdb.Client
	ctx    context.Context
}

type tvEpisode struct {
	result  tmdb.EpisodeDetails
	airDate time.Time
	season  provider.ResponseTVSeason

	provider.ResponseBaseTVEpisode
}

func (c *Client) newTVEpisodeResponse(result tmdb.EpisodeDetails, season provider.ResponseTVSeason, req provider.Request) (*tvEpisodeResponse, error) {
	t, err := newTVEpisode(result, season, req)
	if err != nil {
		return nil, err
	}

	m := &tvEpisodeResponse{
		tvEpisode: t,
		client:    c.client,
		ctx:       c.ctx,
	}
	m.multi = map[string]*tvEpisode{
		req.DestinationLanguage: m.tvEpisode,
	}

	return m, nil
}

func newTVEpisode(result tmdb.EpisodeDetails, season provider.ResponseTVSeason, req provider.Request) (t *tvEpisode, err error) {
	t = &tvEpisode{
		result:                result,
		season:                season,
		ResponseBaseTVEpisode: provider.NewResponseBaseTVEpisode(),
	}
	t.SetRequest(req)

	if result.AirDate != "" {
		// log.Debug().Msgf("parsing tv episode air date: %s", result.AirDate)
		// Parse the first air date in the format "2006-01-02"
		t.airDate, err = time.Parse(time.DateOnly, result.AirDate)
		if err != nil {
			return nil, err
		}
	}

	return t, nil
}

func (r tvEpisode) GetID() int {
	return int(r.result.ID)
}

func (r tvEpisode) GetName() string {
	return r.result.Name
}

func (r tvEpisode) GetDate() time.Time {
	return r.airDate
}

func (r tvEpisode) GetPopularity() int {
	return util.ComputePopularity(-1, r.result.VoteAverage, r.result.VoteCount)
}

func (r tvEpisode) GetEpisodeNumber() int {
	return r.result.EpisodeNumber
}

func (r tvEpisode) GetSeason() provider.ResponseTVSeason {
	return r.season
}

func (r tvEpisode) GetProvider() string {
	return name
}

func (m *tvEpisodeResponse) InLanguage(req provider.Request) (provider.Response, error) {
	if r, ok := m.multi[req.DestinationLanguage]; ok {
		m.tvEpisode = r
	} else {
		languageQuery := buildLanguageQuery(req.DestinationLanguage)
		details, err := m.client.GetTVEpisode(m.ctx, m.GetSeason().GetShow().GetID(), m.GetSeason().GetSeasonNumber(), m.GetID(), languageQuery)
		if err != nil {
			return nil, err
		}

		e, err := newTVEpisode(*details, m.GetSeason(), req)
		if err != nil {
			return nil, err
		}

		m.multi[req.DestinationLanguage] = e
		m.tvEpisode = e
	}

	return m, nil
}
