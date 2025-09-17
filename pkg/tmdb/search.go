package tmdb

import (
	"fmt"
	"strconv"

	"github.com/TheoBrigitte/evansky/pkg/provider"
)

func (c *Client) Search(req provider.Request) (responses []provider.Response, err error) {
	switch req.MediaType {
	case provider.MediaTypeMovie:
		responses, err = c.searchMovies(req)
	case provider.MediaTypeTV:
		responses, err = c.searchTV(req)
	default:
		// If no media type is specified, find most popular between movies and tv shows.
		responses, err = c.searchByPopularity(req)
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
func (c *Client) searchByPopularity(req provider.Request) ([]provider.Response, error) {
	movies, err := c.searchMovies(req)
	if err != nil {
		return nil, err
	}

	tvshows, err := c.searchTV(req)
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
func (c *Client) searchMovies(req provider.Request) ([]provider.Response, error) {
	year := strconv.Itoa(req.Year)

	var additionalQuery = make(map[string]string)
	if year != "" {
		additionalQuery["year"] = year
	}
	movies, err := c.client.GetSearchMovies(req.Query, additionalQuery)
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
func (c *Client) searchTV(req provider.Request) ([]provider.Response, error) {
	if req.TV != nil && req.TV.ID != 0 {
		if req.TV.SeasonID != 0 {
			if req.TV.EpisodeID != 0 {
				// Search for an episode by ids
				return c.getTVEpisode(*req.TV)
			}

			if req.Query != "" {
				// Search for an episode by name
				return c.searchTVEpisode(req)
			}

			// Search for a season by id
			return c.getTVSeason(*req.TV)
		}

		if req.TV.EpisodeID != 0 {
			// Search for an episode by ids
			// TODO: or fail because we don't have season id? or set it to 0?
			return c.getTVEpisode(*req.TV)
		}

		// Search either for a season or an episode.
		return c.searchTVSeasonOrEpisode(req)
	}

	return c.searchTVShow(req)
}

func (c *Client) searchTVShow(req provider.Request) ([]provider.Response, error) {
	year := strconv.Itoa(req.Year)

	var additionalQuery = make(map[string]string)
	if year != "" {
		additionalQuery["year"] = year
	}
	tvshows, err := c.client.GetSearchTVShow(req.Query, additionalQuery)
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

func (c *Client) searchTVSeasonOrEpisode(req provider.Request) ([]provider.Response, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *Client) searchTVEpisode(req provider.Request) ([]provider.Response, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *Client) getTVEpisode(req provider.RequestTV) ([]provider.Response, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *Client) getTVSeason(req provider.RequestTV) ([]provider.Response, error) {
	return nil, fmt.Errorf("not implemented")
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
