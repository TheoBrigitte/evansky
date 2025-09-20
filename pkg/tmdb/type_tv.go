package tmdb

import (
	"fmt"
	"log/slog"
	"time"

	gotmdb "github.com/cyruzin/golang-tmdb"

	"github.com/TheoBrigitte/evansky/pkg/provider"
)

type tvResponse struct {
	provider.ResponseBaseTV

	result       gotmdb.TVShowResult
	firstAirDate time.Time
	// Language indexed seasons cache
	seasons map[string][]provider.ResponseTVSeason

	client *Client
}

func (c *Client) newTVResponse(result gotmdb.TVShowResult) (*tvResponse, error) {
	// Parse the first air date in the format "2006-01-02"
	firstAirDate, err := time.Parse(time.DateOnly, result.FirstAirDate)
	if err != nil {
		return nil, err
	}

	t := &tvResponse{
		ResponseBaseTV: provider.NewResponseBaseTV(),

		result:       result,
		firstAirDate: firstAirDate,
		seasons:      make(map[string][]provider.ResponseTVSeason),

		client: c,
	}
	return t, nil
}

func (r tvResponse) GetID() int {
	return int(r.result.ID)
}

func (r tvResponse) GetName() string {
	return r.result.Name
}

func (r tvResponse) GetDate() time.Time {
	return r.firstAirDate
}

func (r tvResponse) GetPopularity() int {
	return computePopularity(r.result.Popularity, r.result.VoteAverage, r.result.VoteCount)
}

func (r tvResponse) GetSeasons(req provider.Request) ([]provider.ResponseTVSeason, error) {
	slog.Debug("get seasons", "show_id", r.GetID(), "language", req.Language)
	if s, ok := r.seasons[req.Language]; ok {
		return s, nil
	}

	languageQuery := buildLanguageQuery(req)
	resp, err := r.client.client.GetTVDetails(r.GetID(), languageQuery)
	if err != nil {
		return nil, err
	}

	seasons := make([]provider.ResponseTVSeason, 0, len(resp.Seasons))
	for _, s := range resp.Seasons {
		season, err := r.client.client.GetTVSeasonDetails(r.GetID(), s.SeasonNumber, languageQuery)
		if err != nil {
			slog.Warn("GetSeasons: failed to get season details", "show_id", r.GetID(), "season_number", s.SeasonNumber, "error", err)
			continue
		}
		r, err := r.client.newTVSeasonResponse(*season, r)
		if err != nil {
			continue
		}
		seasons = append(seasons, r)
	}

	if len(seasons) == 0 {
		return nil, fmt.Errorf("no season found for show %d", r.GetID())
	}

	r.seasons[req.Language] = seasons

	return seasons, nil
}

func (r tvResponse) GetSeason(seasonNumber int, req provider.Request) (provider.ResponseTVSeason, error) {
	slog.Debug("get season", "show_id", r.GetID(), "season_number", seasonNumber, "language", req.Language)
	seasons, err := r.GetSeasons(req)
	if err != nil {
		return nil, err
	}

	for _, s := range seasons {
		if s.GetSeasonNumber() == seasonNumber {
			return s, nil
		}
	}

	return nil, fmt.Errorf("season %d not found for show %d", seasonNumber, r.GetID())
}
