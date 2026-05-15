package tmdb

import (
	"errors"
	"strconv"

	"github.com/golusoris/goenvoy/metadata/video/tmdb"
	"github.com/rs/zerolog/log"

	"github.com/TheoBrigitte/evansky/pkg/provider"
)

// SearchMovie search for movies using query and year (if provided).
// see: https://developer.themoviedb.org/reference/search-movie
func (c *Client) SearchMovie(req provider.Request) (provider.ResponseMovie, float64, error) {
	movies, err := c.searchMovie(req)
	if err != nil {
		if !errors.Is(err, provider.ErrNoResult) {
			return nil, 0, err
		}

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

func (c *Client) searchMovie(req provider.Request) (*tmdb.PaginatedResult[tmdb.MovieResult], error) {
	query := buildAdditionalQuery(req)
	log.Debug().Str("query", query).Any("language", req.QueryLanguage).Msg("searching movie")
	movies, err := c.client.SearchMovies(c.ctx, query, req.QueryLanguage, 1)
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
	if err != nil {
		if !errors.Is(err, provider.ErrNoResult) {
			return nil, 0, err
		}

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

func (c *Client) searchTV(req provider.Request) (*tmdb.PaginatedResult[tmdb.TVResult], error) {
	query := buildAdditionalQuery(req)
	log.Debug().Str("query", query).Any("language", req.QueryLanguage).Msg("searching tv")
	tvshows, err := c.client.SearchTV(c.ctx, query, req.QueryLanguage, 1)
	if err != nil {
		return nil, err
	}
	if len(tvshows.Results) <= 0 {
		return nil, provider.ErrNoResult
	}

	return tvshows, nil
}

func buildAdditionalQuery(req provider.Request) string {
	q := req.Query
	if req.Year != 0 {
		q += " " + strconv.Itoa(req.Year)
	}
	return q
}

func buildLanguageQuery(language string) string {
	return language
}
