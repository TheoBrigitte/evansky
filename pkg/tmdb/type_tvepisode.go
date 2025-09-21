package tmdb

import (
	"time"

	gotmdb "github.com/cyruzin/golang-tmdb"

	"github.com/TheoBrigitte/evansky/pkg/provider"
)

type tvEpisodeResponse struct {
	*tvEpisode
	multi  map[string]*tvEpisode
	client *gotmdb.Client

	provider.ResponseBaseTVEpisode
}

type tvEpisode struct {
	result  gotmdb.TVEpisodeDetails
	airDate time.Time
	season  provider.ResponseTVSeason
}

func (c *Client) newTVEpisodeResponse(result gotmdb.TVEpisodeDetails, season provider.ResponseTVSeason, lang string) (*tvEpisodeResponse, error) {
	t, err := newTVEpisode(result, season)
	if err != nil {
		return nil, err
	}

	m := &tvEpisodeResponse{
		tvEpisode:             t,
		client:                c.client,
		ResponseBaseTVEpisode: provider.NewResponseBaseTVEpisode(),
	}
	m.multi = map[string]*tvEpisode{
		lang: m.tvEpisode,
	}

	return m, nil
}

func newTVEpisode(result gotmdb.TVEpisodeDetails, season provider.ResponseTVSeason) (*tvEpisode, error) {
	// Parse the first air date in the format "2006-01-02"
	airDate, err := time.Parse(time.DateOnly, result.AirDate)
	if err != nil {
		return nil, err
	}

	t := &tvEpisode{
		result:  result,
		airDate: airDate,
		season:  season,
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
	return computePopularity(-1, r.result.VoteAverage, r.result.VoteCount)
}

func (r tvEpisode) GetEpisodeNumber() int {
	return r.result.EpisodeNumber
}

func (r tvEpisode) GetSeason() provider.ResponseTVSeason {
	return r.season
}

func (m *tvEpisodeResponse) InLanguage(req provider.Request) (provider.Response, error) {
	if r, ok := m.multi[req.Language]; ok {
		m.tvEpisode = r
	} else {

		languageQuery := buildLanguageQuery(req)
		details, err := m.client.GetTVEpisodeDetails(m.GetSeason().GetShow().GetID(), m.GetSeason().GetSeasonNumber(), m.GetID(), languageQuery)
		if err != nil {
			return nil, err
		}

		e, err := newTVEpisode(*details, m.GetSeason())
		if err != nil {
			return nil, err
		}

		m.multi[req.Language] = e
		m.tvEpisode = e
	}

	return m, nil
}
