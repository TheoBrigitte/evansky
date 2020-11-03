package scan

import (
	"fmt"

	"github.com/TheoBrigitte/evansky/pkg/movie"
	"github.com/TheoBrigitte/evansky/pkg/tmdb"
)

type Scanner struct {
	client *tmdb.Client
	noAPI  bool
}

type Config struct {
	APIKey string
	NoAPI  bool
}

type Results struct {
	Total    int                    `json:"total"`
	Found    int                    `json:"found"`
	Failures int                    `json:"failures"`
	Results  map[string]movie.Movie `json:"results"`
	Failed   map[string]movie.Movie `json:"failed"`
}

func NewResults() *Results {
	r := &Results{}
	r.Results = make(map[string]movie.Movie)
	r.Failed = make(map[string]movie.Movie)

	return r
}

func (r *Results) IsComplete() bool {
	return r.Found == r.Total
}

func (r *Results) CompletePercentage() string {
	return fmt.Sprintf("%d%%", r.Found*100.0/r.Total)
}
