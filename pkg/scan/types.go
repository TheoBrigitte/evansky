package scan

import (
	"fmt"

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

type Result struct {
	Path     string `json:"path"`
	ID       int64  `json:"id"`
	Language string `json:"language"`
	IsDir    bool   `json:"isDir"`
}

type Results struct {
	Total   int               `json:"total"`
	Found   int               `json:"found"`
	Results map[string]Result `json:"results"`
}

func NewResults() *Results {
	r := &Results{}
	r.Results = make(map[string]Result)

	return r
}

func (r *Results) IsComplete() bool {
	return r.Found == r.Total
}

func (r *Results) CompletePercentage() string {
	return fmt.Sprintf("%d%%", r.Found*100.0/r.Total)
}
