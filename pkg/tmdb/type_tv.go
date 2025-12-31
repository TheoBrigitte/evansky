package tmdb

import (
	"fmt"
	"time"

	gotmdb "github.com/cyruzin/golang-tmdb"
	"github.com/rs/zerolog/log"

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

	if result.FirstAirDate != "" {
		// log.Debug().Msg("parsing tv first air date: " + result.FirstAirDate)
		// Parse the first air date in the format "2006-01-02"
		firstAirDate, err := time.Parse(time.DateOnly, result.FirstAirDate)
		if err != nil {
			return err
		}
		m.tv.firstAirDate = firstAirDate
	}

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
	log.Debug().Msgf("TV show %d seasons loaded: %d", m.GetID(), len(m.seasons))
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

func tvshowByClosestYear(year int, tvshows []gotmdb.TVShowResult) gotmdb.TVShowResult {
	if year == 0 {
		return tvshows[0]
	}

	var bestScore float64 = 0
	var closestMatch gotmdb.TVShowResult

	for index, t := range tvshows {
		date, err := time.Parse(time.DateOnly, t.FirstAirDate)
		if err != nil {
			log.Warn().Err(err).Msgf("failed to parse FirstAirDate: %s", t.FirstAirDate)
			continue
		}
		score := computeClosetYearScore(year, date.Year(), index)
		log.Debug().Msgf("comparing tv shows %s tmdbid=%d date=%s score=%f", t.Name, t.ID, t.FirstAirDate, score)

		if bestScore == 0 || score < float64(bestScore) {
			bestScore = score
			closestMatch = t
		}
	}

	return closestMatch
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
