package tmdb

import (
	"strconv"

	"github.com/TheoBrigitte/evansky/pkg/provider"
	"github.com/rs/zerolog/log"
)

// SearchMovie search for movies using query and year (if provided).
// see: https://developer.themoviedb.org/reference/search-movie
func (c *Client) SearchMovie(req provider.Request) (provider.ResponseMovie, error) {
	additionalQuery := buildAdditionalQuery(req)
	log.Debug().Str("query", req.Query).Any("additionalQuery", additionalQuery).Msg("searching movie")
	movies, err := c.client.GetSearchMovies(req.Query, additionalQuery)
	if err != nil {
		return nil, err
	}
	if len(movies.Results) <= 0 {
		return nil, provider.ErrNoResult
	}

	return c.newMovieResponse(movies.Results[0], req.Language)
}

// SearchTV search for tv shows using query and year (if provided).
// see: https://developer.themoviedb.org/reference/search-tv
func (c *Client) SearchTV(req provider.Request) (provider.ResponseTV, error) {
	additionalQuery := buildAdditionalQuery(req)
	log.Debug().Str("query", req.Query).Any("additionalQuery", additionalQuery).Msg("searching tv")
	tvshows, err := c.client.GetSearchTVShow(req.Query, additionalQuery)
	if err != nil {
		return nil, err
	}
	if len(tvshows.Results) <= 0 {
		return nil, provider.ErrNoResult
	}

	return c.newTVResponse(tvshows.Results[0], req)
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
	additionalQuery := make(map[string]string)
	if req.Year != 0 {
		additionalQuery["year"] = strconv.Itoa(req.Year)
	}
	if req.Language != "" {
		additionalQuery["language"] = req.Language
	}
	return additionalQuery
}

func buildLanguageQuery(req provider.Request) map[string]string {
	additionalQuery := make(map[string]string)
	if req.Language != "" {
		additionalQuery["language"] = req.Language
	}
	return additionalQuery
}
