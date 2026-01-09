package tmdb

import (
	"errors"
	"strconv"

	gotmdb "github.com/cyruzin/golang-tmdb"
	"github.com/rs/zerolog/log"

	"github.com/TheoBrigitte/evansky/pkg/provider"
)

// SearchMovie search for movies using query and year (if provided).
// see: https://developer.themoviedb.org/reference/search-movie
func (c *Client) SearchMovie(req provider.Request) (provider.ResponseMovie, float64, error) {
	movies, err := c.searchMovie(req)
	if err != nil && errors.Is(err, provider.ErrNoResult) {
		// Try again without year and language filters
		req.Year = 0
		req.QueryLanguage = ""
		movies, err = c.searchMovie(req)
		if err != nil {
			return nil, 0, err
		}
	}

	result, score := movieByClosestYear(req.Query, req.Year, movies.Results)
	resp, err := c.newMovieResponse(result, req.QueryLanguage)
	return resp, score, err
}

func (c *Client) searchMovie(req provider.Request) (*gotmdb.SearchMovies, error) {
	additionalQuery := buildAdditionalQuery(req)
	log.Debug().Str("query", req.Query).Any("additionalQuery", additionalQuery).Msg("searching movie")
	movies, err := c.client.GetSearchMovies(req.Query, additionalQuery)
	if err != nil {
		return nil, err
	}
	if len(movies.Results) <= 0 {
		return nil, provider.ErrNoResult
	}

	return movies, nil
}

// SearchTV search for tv shows using query and year (if provided).
// see: https://developer.themoviedb.org/reference/search-tv
func (c *Client) SearchTV(req provider.Request) (provider.ResponseTV, float64, error) {
	tvshows, err := c.searchTV(req)
	if err != nil && errors.Is(err, provider.ErrNoResult) {
		// Try again without year and language filters
		req.Year = 0
		req.QueryLanguage = ""
		tvshows, err = c.searchTV(req)
		if err != nil {
			return nil, 0, err
		}
	}

	result, score := tvshowByClosestYear(req.Query, req.Year, tvshows.Results)
	resp, err := c.newTVResponse(result, req)
	return resp, score, err
}

func (c *Client) searchTV(req provider.Request) (*gotmdb.SearchTVShows, error) {
	additionalQuery := buildAdditionalQuery(req)
	log.Debug().Str("query", req.Query).Any("additionalQuery", additionalQuery).Msg("searching tv")
	tvshows, err := c.client.GetSearchTVShow(req.Query, additionalQuery)
	if err != nil {
		return nil, err
	}
	if len(tvshows.Results) <= 0 {
		return nil, provider.ErrNoResult
	}

	return tvshows, nil
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
	if req.QueryLanguage != "" {
		additionalQuery["language"] = req.QueryLanguage
	}
	return additionalQuery
}

func buildLanguageQuery(language string) map[string]string {
	additionalQuery := make(map[string]string)
	if language != "" {
		additionalQuery["language"] = language
	}
	return additionalQuery
}
