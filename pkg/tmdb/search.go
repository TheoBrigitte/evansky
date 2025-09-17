package tmdb

import (
	"fmt"
	"strconv"

	"github.com/TheoBrigitte/evansky/pkg/provider"
)

func (c *Client) Search(req provider.Request) (responses []provider.Response, err error) {
	year := strconv.Itoa(req.Year)

	switch req.MediaType {
	case provider.MediaTypeMovie:
		responses, err = c.searchMovies(req.Query, year)
	case provider.MediaTypeTV:
		responses, err = c.searchTV(req.Query, year)
	default:
		// If no media type is specified, find most popular between movies and tv shows.
		responses, err = c.searchByPopularity(req.Query, year)
	}

	if err != nil {
		return nil, err
	}

	if len(responses) == 0 {
		return nil, fmt.Errorf("no result")
	}

	return responses, nil
}

// searchByPopularity search most popular results between movies and tv shows.
// If one of the two types has no result, return the other type results.
// If both types have results, compare the popularity of the first result of each type and return
func (c *Client) searchByPopularity(query, year string) ([]provider.Response, error) {
	movies, err := c.searchMovies(query, year)
	if err != nil {
		return nil, err
	}

	tvshows, err := c.searchTV(query, year)
	if err != nil {
		return nil, err
	}

	var responses []provider.Response
	switch {
	case len(movies) == 0:
		responses = append(responses, tvshows...)
	case len(tvshows) == 0:
		responses = append(responses, movies...)
	case movies[0].GetPopularity() >= tvshows[0].GetPopularity():
		responses = append(responses, movies...)
	default:
		responses = append(responses, tvshows...)
	}

	return responses, nil
}

// searchMovies search for movies using query and year (if provided).
// see: https://developer.themoviedb.org/reference/search-movie
func (c *Client) searchMovies(query, year string) ([]provider.Response, error) {
	var additionalQuery = make(map[string]string)
	if year != "" {
		additionalQuery["year"] = year
	}
	movies, err := c.client.GetSearchMovies(query, additionalQuery)
	if err != nil {
		return nil, err
	}

	responses := make([]provider.Response, 0, len(movies.Results))
	for _, result := range movies.Results {
		r, err := newMovieResponse(result)
		if err != nil {
			return nil, err
		}
		responses = append(responses, r)
	}

	return responses, nil
}

// searchTV search for tv shows using query and year (if provided).
// see: https://developer.themoviedb.org/reference/search-tv
func (c *Client) searchTV(query, year string) ([]provider.Response, error) {
	var additionalQuery = make(map[string]string)
	if year != "" {
		additionalQuery["year"] = year
	}
	tvshows, err := c.client.GetSearchTVShow(query, additionalQuery)
	if err != nil {
		return nil, err
	}

	responses := make([]provider.Response, 0, len(tvshows.Results))
	for _, result := range tvshows.Results {
		r, err := newTVResponse(result)
		if err != nil {
			return nil, err
		}
		responses = append(responses, r)
	}

	return responses, nil
}

// searchMulti search for multi media using query.
// see: https://developer.themoviedb.org/reference/search-multi
func (c *Client) searchMulti(query, year string) ([]provider.Response, error) {
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

	return responses, nil
}
