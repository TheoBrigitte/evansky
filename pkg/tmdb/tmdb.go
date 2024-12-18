package tmdb

import (
	gotmdb "github.com/cyruzin/golang-tmdb"
)

// New return a new tmdb client.
func New(config Config) (*Client, error) {
	tmdbClient, err := gotmdb.Init(config.APIKey)
	if err != nil {
		return nil, err
	}

	c := &Client{
		client: tmdbClient,
	}

	return c, nil
}

// GetMovies search for movies using query and year (if provided).
// see: https://developers.themoviedb.org/3/search/search-movies
func (c *Client) GetMovies(query, year string) (*gotmdb.SearchMovies, error) {
	var additionalQuery = make(map[string]string)
	if year != "" {
		additionalQuery["year"] = year
	}
	movies, err := c.client.GetSearchMovies(query, additionalQuery)
	if err != nil {
		return nil, err
	}

	return movies, nil
}

// GetMovies search for movies using query and year (if provided).
// see: https://developers.themoviedb.org/3/search/search-movies
func (c *Client) GetMulti(query, year string) (*gotmdb.SearchMulti, error) {
	var additionalQuery = make(map[string]string)
	if year != "" {
		additionalQuery["year"] = year
	}
	multi, err := c.client.GetSearchMulti(query, additionalQuery)
	if err != nil {
		return nil, err
	}

	return multi, nil
}
