package tmdb

import (
	"errors"
	"fmt"
	"strconv"

	gotmdb "github.com/cyruzin/golang-tmdb"
	"github.com/gogf/gf/v2/text/gstr"

	"github.com/TheoBrigitte/evansky/pkg/provider"
)

func (c *Client) Search(req provider.Request) (response provider.Response, err error) {
	switch req.MediaType {
	case provider.MediaTypeMovie:
		response, err = c.searchMovies(req)
	case provider.MediaTypeTV:
		response, err = c.searchTV(req)
	default:
		// If no media type is specified, find most popular between movies and tv shows.
		response, err = c.searchByPopularity(req)
	}

	if err != nil {
		return nil, err
	}

	return response, nil
}

// searchByPopularity search most popular results between movies and tv shows.
// If one of the two types has no result, return the other type results.
// If both types have results, compare the popularity of the first result of each type and return
func (c *Client) searchByPopularity(req provider.Request) (provider.Response, error) {
	movie, err := c.searchMovies(req)
	if err != nil {
		return nil, err
	}

	tvshow, err := c.searchTV(req)
	if err != nil {
		return nil, err
	}

	if movie.GetPopularity() >= tvshow.GetPopularity() {
		return movie, nil
	}

	return tvshow, nil
}

// searchMovies search for movies using query and year (if provided).
// see: https://developer.themoviedb.org/reference/search-movie
func (c *Client) searchMovies(req provider.Request) (provider.Response, error) {
	additionalQuery := buildAdditionalQuery(req)
	movies, err := c.client.GetSearchMovies(req.Query, additionalQuery)
	if err != nil {
		return nil, err
	}

	return makeResponse(movies.Results, newMovieResponse)
}

// searchTV search for tv shows using query and year (if provided).
// see: https://developer.themoviedb.org/reference/search-tv
func (c *Client) searchTV(req provider.Request) (provider.Response, error) {
	if req.TV != nil && req.TV.ID != 0 {
		if req.TV.SeasonNumber != 0 {
			if req.TV.EpisodeNumber != 0 {
				// When season and episode numbers are provided
				// Get the episode by id
				return c.getTVEpisode(req)
			}

			if req.Query == "" {
				// When only season number is provided but no query
				// Get the season by id
				return c.getTVSeason(req)
			}

			// When season number and query are provided
			// Search for an episode by name
			return c.searchTVEpisode(req)
		}

		if req.TV.EpisodeNumber != 0 {
			// When only episode number is provided
			// TODO: get episode by id
			// TODO: or fail because we don't have season id? or set it to 0?
			// TODO: or use query and number o search for the episode?
			return nil, fmt.Errorf("searching by episode number without season number is not supported")
		}

		// When no season or episode number is provided
		// Search either for a season or an episode
		return c.searchTVSeasonOrEpisode(req)
	}

	return c.searchTVShow(req)
}

func (c *Client) searchTVShow(req provider.Request) (provider.Response, error) {
	additionalQuery := buildAdditionalQuery(req)
	tvshows, err := c.client.GetSearchTVShow(req.Query, additionalQuery)
	if err != nil {
		return nil, err
	}

	return makeResponse(tvshows.Results, newTVResponse)
}

func (c *Client) searchTVSeasonOrEpisode(req provider.Request) (provider.Response, error) {
	// TODO: after implementing the response as part of the request, try to search in the previous response if any
	if req.TV == nil || req.TV.ID == 0 {
		return nil, fmt.Errorf("no tv id provided")
	}
	if req.Query == "" {
		return nil, fmt.Errorf("no query provided")
	}

	languageQuery := buildLanguageQuery(req)
	show, err := c.client.GetTVDetails(req.TV.ID, languageQuery)
	if err != nil {
		return nil, err
	}

	var bestMatch provider.Response
	var bestScore int = -1
	//seasons := make([]gotmdb.TVSeason, 0, len(show.Seasons))
	for _, s := range show.Seasons {
		season, err := c.client.GetTVSeasonDetails(req.TV.ID, s.SeasonNumber, languageQuery)
		if err != nil {
			return nil, err
		}
		seasonScore := gstr.Levenshtein(req.Query, season.Name, 1, 1, 1)
		if bestScore == -1 || seasonScore < bestScore {
			bestScore = seasonScore
			bestMatch, err = newTVSeasonResponse(*season, req.TV.ID)
			if err != nil {
				return nil, err
			}
		}

		for _, e := range season.Episodes {
			episodeScore := gstr.Levenshtein(req.Query, e.Name, 1, 1, 1)
			if bestScore == -1 || episodeScore < bestScore {
				bestScore = episodeScore
				// This is a dirty way to convert an episode to gotmdb.TVEpisodeDetails, as season.Episodes has no concrete type.
				var r gotmdb.TVEpisodeDetails
				r.AirDate = e.AirDate
				r.EpisodeNumber = e.EpisodeNumber
				r.ID = e.ID
				r.Name = e.Name
				r.Overview = e.Overview
				r.ProductionCode = e.ProductionCode
				r.Runtime = e.Runtime
				r.SeasonNumber = e.SeasonNumber
				r.StillPath = e.StillPath
				r.VoteMetrics = e.VoteMetrics
				r.Crew = e.Crew
				r.GuestStars = e.GuestStars
				bestMatch, err = newTVEpisodeResponse(r, req.TV.ID)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	if bestMatch == nil {
		return nil, fmt.Errorf("no result")
	}

	return bestMatch, nil
}

func (c *Client) searchTVEpisode(req provider.Request) (provider.Response, error) {
	languageQuery := buildLanguageQuery(req)
	season, err := c.client.GetTVSeasonDetails(req.TV.ID, req.TV.SeasonNumber, languageQuery)
	if err != nil {
		return nil, err
	}

	var bestMatch provider.Response
	var bestScore int = -1
	for _, e := range season.Episodes {
		episodeScore := gstr.Levenshtein(req.Query, e.Name, 1, 1, 1)
		if bestScore == -1 || episodeScore < bestScore {
			bestScore = episodeScore
			// This is a dirty way to convert an episode to gotmdb.TVEpisodeDetails, as season.Episodes has no concrete type.
			var r gotmdb.TVEpisodeDetails
			r.AirDate = e.AirDate
			r.EpisodeNumber = e.EpisodeNumber
			r.ID = e.ID
			r.Name = e.Name
			r.Overview = e.Overview
			r.ProductionCode = e.ProductionCode
			r.Runtime = e.Runtime
			r.SeasonNumber = e.SeasonNumber
			r.StillPath = e.StillPath
			r.VoteMetrics = e.VoteMetrics
			r.Crew = e.Crew
			r.GuestStars = e.GuestStars
			bestMatch, err = newTVEpisodeResponse(r, req.TV.ID)
			if err != nil {
				return nil, err
			}
		}
	}

	if bestMatch == nil {
		return nil, errors.New("no result")
	}

	return bestMatch, nil
}

func (c *Client) getTVSeason(req provider.Request) (provider.Response, error) {
	if req.TV == nil {
		return nil, fmt.Errorf("no tv info provided")
	}

	languageQuery := buildLanguageQuery(req)
	resp, err := c.client.GetTVSeasonDetails(req.TV.ID, req.TV.SeasonNumber, languageQuery)
	if err != nil {
		return nil, err
	}

	//return makeResponse([]gotmdb.TVSeasonDetails{*resp}, newTVSeasonResponse)

	return newTVSeasonResponse(*resp, req.TV.ID)
}

func (c *Client) getTVEpisode(req provider.Request) (provider.Response, error) {
	if req.TV == nil {
		return nil, fmt.Errorf("no tv info provided")
	}

	languageQuery := buildLanguageQuery(req)
	resp, err := c.client.GetTVEpisodeDetails(req.TV.ID, req.TV.SeasonNumber, req.TV.EpisodeNumber, languageQuery)
	if err != nil {
		return nil, err
	}

	//if resp.ID == 0 {
	//	return nil, fmt.Errorf("no result")
	//}

	//return makeResponse([]gotmdb.TVEpisodeDetails{*resp}, newTVEpisodeResponse)

	return newTVEpisodeResponse(*resp, req.TV.ID)
}

// searchMulti search for multi media using query.
// see: https://developer.themoviedb.org/reference/search-multi
func (c *Client) searchMulti(query, year string) (provider.Response, error) {
	multi, err := c.client.GetSearchMulti(query, nil)
	if err != nil {
		return nil, err
	}

	responses := make([]provider.Response, 0, len(multi.Results))
	//for _, result := range multi.Results {
	//	r := multiResponse{
	//		ID:    result.ID,
	//		Title: result.Title,
	//		Name:  result.Name,
	//		//MediaType:        result.MediaType,
	//		OriginalLanguage: result.OriginalLanguage,
	//		OriginalName:     result.OriginalName,
	//		OriginalTitle:    result.OriginalTitle,
	//		ReleaseDate:      result.ReleaseDate,
	//		FirstAirDate:     result.FirstAirDate,
	//		Popularity:       result.Popularity,
	//	}
	//	responses = append(responses, r)
	//}

	return responses[0], nil
}

func buildAdditionalQuery(req provider.Request) map[string]string {
	var additionalQuery = make(map[string]string)
	if req.Year != 0 {
		additionalQuery["year"] = strconv.Itoa(req.Year)
	}
	if req.Language != "" {
		additionalQuery["language"] = req.Language
	}
	return additionalQuery
}

func buildLanguageQuery(req provider.Request) map[string]string {
	var additionalQuery = make(map[string]string)
	if req.Language != "" {
		additionalQuery["language"] = req.Language
	}
	return additionalQuery
}
