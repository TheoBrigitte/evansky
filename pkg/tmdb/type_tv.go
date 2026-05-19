package tmdb

import (
	"fmt"
	"time"

	"github.com/golusoris/goenvoy/metadata/video/tmdb"
	"github.com/rs/zerolog/log"

	"github.com/TheoBrigitte/evansky/pkg/provider"
	"github.com/TheoBrigitte/evansky/pkg/util"
)

type tvResponse struct {
	*tv
	multi  map[string]*tv
	client *Client
}

type tv struct {
	result       tmdb.TVResult
	firstAirDate time.Time
	// Language indexed seasons cache
	seasons []provider.ResponseTVSeason

	provider.ResponseBaseTV
}

func (c *Client) newTVResponse(result tmdb.TVResult, req provider.Request) (*tvResponse, error) {
	m := &tvResponse{
		multi:  make(map[string]*tv),
		client: c,
	}

	err := m.newTv(result, req)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (m *tvResponse) newTv(result tmdb.TVResult, req provider.Request) error {
	m.tv = &tv{
		result:         result,
		ResponseBaseTV: provider.NewResponseBaseTV(),
	}
	m.SetRequest(req)

	if result.FirstAirDate != "" {
		// log.Debug().Msg("parsing tv first air date: " + result.FirstAirDate)
		// Parse the first air date in the format "2006-01-02"
		firstAirDate, err := time.Parse(time.DateOnly, result.FirstAirDate)
		if err != nil {
			return err
		}
		m.firstAirDate = firstAirDate
	}

	languageQuery := buildLanguageQuery(req.DestinationLanguage)
	resp, err := m.client.client.GetTV(m.client.ctx, m.GetID(), languageQuery)
	if err != nil {
		return err
	}

	seasons := make([]provider.ResponseTVSeason, 0, len(resp.Seasons))
	for _, s := range resp.Seasons {
		season, err := m.client.client.GetTVSeason(m.client.ctx, m.GetID(), s.SeasonNumber, languageQuery)
		if err != nil {
			return err
		}

		r, err := m.client.newTVSeasonResponse(*season, m, req)
		if err != nil {
			return err
		}
		seasons = append(seasons, r)
	}
	m.seasons = seasons
	log.Debug().Msgf("TV show %d seasons loaded: %d", m.GetID(), len(m.seasons))
	m.multi[req.DestinationLanguage] = m.tv

	return nil
}

func (r tv) GetID() int {
	return int(r.result.ID)
}

func (r tv) GetName() string {
	return r.result.Name
}

func (r tv) GetDate() time.Time {
	return r.firstAirDate
}

func (r tv) GetPopularity() int {
	return util.ComputePopularity(r.result.Popularity, r.result.VoteAverage, r.result.VoteCount)
}

func (r tv) GetProvider() string {
	return name
}

func (r *tv) GetSeasons() []provider.ResponseTVSeason {
	// slog.Debug("get seasons", "show_id", r.GetID(), "seasons", len(r.seasons))
	return r.seasons
}

func (r tv) GetSeason(seasonNumber int) (provider.ResponseTVSeason, error) {
	// slog.Debug("get season", "show_id", r.GetID(), "season_number", seasonNumber)

	for _, s := range r.GetSeasons() {
		if s.GetSeasonNumber() == seasonNumber {
			return s, nil
		}
	}

	return nil, fmt.Errorf("season %d not found for show %d", seasonNumber, r.GetID())
}

func (m *tvResponse) InLanguage(req provider.Request) (provider.Response, error) {
	if r, ok := m.multi[req.DestinationLanguage]; ok {
		m.tv = r
	} else {
		languageQuery := buildLanguageQuery(req.DestinationLanguage)
		details, err := m.client.client.GetTV(m.client.ctx, m.GetID(), languageQuery)
		if err != nil {
			return nil, err
		}

		result := tmdb.TVResult{
			ID:               details.ID,
			Name:             details.Name,
			OriginalName:     details.OriginalName,
			OriginalLanguage: details.OriginalLanguage,
			Overview:         details.Overview,
			FirstAirDate:     details.FirstAirDate,
			PosterPath:       details.PosterPath,
			BackdropPath:     details.BackdropPath,
			Popularity:       details.Popularity,
			OriginCountry:    details.OriginCountry,
			VoteAverage:      details.VoteAverage,
			VoteCount:        details.VoteCount,
		}

		err = m.newTv(result, req)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}
