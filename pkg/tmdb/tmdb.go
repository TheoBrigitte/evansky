package tmdb

import (
	"fmt"
	"strconv"

	gotmdb "github.com/cyruzin/golang-tmdb"
	"github.com/spf13/pflag"

	"github.com/TheoBrigitte/evansky/pkg/provider"
)

// New return a new tmdb client.
func New(flags *pflag.FlagSet) (provider.Interface, error) {
	// Validate api key early to catch error before Init.
	if apiKey == "" {
		return nil, fmt.Errorf("--%s is required", apiKeyFlag)
	}

	tmdbClient, err := gotmdb.Init(apiKey)
	if err != nil {
		return nil, err
	}
	tmdbClient.SetClientConfig(newClient(&clientOptions{
		ttl: cacheTTL,
	}))

	c := &Client{
		client: tmdbClient,
	}

	return c, nil
}

func (c *Client) Name() string {
	return name
}

func (c *Client) Search(req provider.Request) ([]provider.Response, error) {
	multi, err := c.GetMulti(req.Query, strconv.Itoa(req.Year))
	if err != nil {
		return nil, err
	}

	responses := make([]provider.Response, 0, len(multi.Results))
	for _, result := range multi.Results {
		r := response{
			ID:               result.ID,
			Title:            result.Title,
			Name:             result.Name,
			MediaType:        result.MediaType,
			OriginalLanguage: result.OriginalLanguage,
			OriginalName:     result.OriginalName,
			OriginalTitle:    result.OriginalTitle,
			ReleaseDate:      result.ReleaseDate,
			FirstAirDate:     result.FirstAirDate,
		}
		responses = append(responses, r)
	}

	if len(responses) == 0 {
		return nil, fmt.Errorf("no result")
	}

	return responses, nil
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
