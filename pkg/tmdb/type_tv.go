package tmdb

import (
	"fmt"
	"log/slog"
	"time"

	gotmdb "github.com/cyruzin/golang-tmdb"

	"github.com/TheoBrigitte/evansky/pkg/provider"
)

type tvResponse struct {
	*tv
	multi  map[string]*tv
	client *Client

	provider.ResponseBaseTV
}

type tv struct {
	result       gotmdb.TVShowResult
	firstAirDate time.Time
	// Language indexed seasons cache
	seasons []provider.ResponseTVSeason
}

func (c *Client) newTVResponse(result gotmdb.TVShowResult, req provider.Request) (*tvResponse, error) {
	m := &tvResponse{
		multi:          make(map[string]*tv),
		client:         c,
		ResponseBaseTV: provider.NewResponseBaseTV(),
	}

	err := m.init(result, req)
	if err != nil {
		return nil, err
	}

	return m, nil
}
func (m *tvResponse) init(result gotmdb.TVShowResult, req provider.Request) error {
	m.tv = &tv{
		result: result,
	}

	// Parse the first air date in the format "2006-01-02"
	firstAirDate, err := time.Parse(time.DateOnly, result.FirstAirDate)
	if err != nil {
		return err
	}
	m.tv.firstAirDate = firstAirDate

	languageQuery := buildLanguageQuery(req)
	resp, err := m.client.client.GetTVDetails(m.GetID(), languageQuery)
	if err != nil {
		return err
	}

	seasons := make([]provider.ResponseTVSeason, 0, len(resp.Seasons))
	for _, s := range resp.Seasons {
		season, err := m.client.client.GetTVSeasonDetails(m.GetID(), s.SeasonNumber, languageQuery)
		if err != nil {
			return err
		}

		r, err := m.client.newTVSeasonResponse(*season, m, req)
		if err != nil {
			return err
		}
		seasons = append(seasons, r)
	}
	m.tv.seasons = seasons
	slog.Debug("tv show seasons loaded", "show_id", m.GetID(), "seasons", len(m.seasons))
	m.multi[req.Language] = m.tv

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
	return computePopularity(r.result.Popularity, r.result.VoteAverage, r.result.VoteCount)
}

func (r *tv) GetSeasons() []provider.ResponseTVSeason {
	//slog.Debug("get seasons", "show_id", r.GetID(), "seasons", len(r.seasons))
	return r.seasons
}

func (r tv) GetSeason(seasonNumber int) (provider.ResponseTVSeason, error) {
	//slog.Debug("get season", "show_id", r.GetID(), "season_number", seasonNumber)

	for _, s := range r.GetSeasons() {
		if s.GetSeasonNumber() == seasonNumber {
			return s, nil
		}
	}

	return nil, fmt.Errorf("season %d not found for show %d", seasonNumber, r.GetID())
}

func (m *tvResponse) InLanguage(req provider.Request) (provider.Response, error) {
	if r, ok := m.multi[req.Language]; ok {
		m.tv = r
	} else {

		languageQuery := buildLanguageQuery(req)
		details, err := m.client.client.GetTVDetails(m.GetID(), languageQuery)
		if err != nil {
			return nil, err
		}

		result := gotmdb.TVShowResult{
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
			VoteMetrics:      details.VoteMetrics,
		}

		err = m.init(result, req)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}
