package scan

import (
	"errors"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/TheoBrigitte/evansky/pkg/movie"
	"github.com/TheoBrigitte/evansky/pkg/parser"
	"github.com/TheoBrigitte/evansky/pkg/tmdb"
)

func New(config Config) (*Scanner, error) {
	var err error
	var client *tmdb.Client
	{
		if !config.NoAPI {
			c := tmdb.Config{
				APIKey: config.APIKey,
			}
			client, err = tmdb.New(c)
			if err != nil {
				return nil, err
			}
		}
	}

	s := &Scanner{
		client: client,
		noAPI:  config.NoAPI,
	}

	return s, nil
}

func (s *Scanner) Scan(files []os.FileInfo, interactive bool, current *Results) (*Results, error) {
	results := NewResults()
	if current != nil {
		results = current
	}
	results.Total = len(files)

	log.Debugf("scanning %d file(s)", results.Total)

	for _, f := range files {
		if _, ok := results.Results[f.Name()]; !ok {
			info, err := parser.Parse(f.Name())
			if err != nil {
				return nil, err
			}

			var m *movie.Movie
			if !s.noAPI {
				m, err = s.searchMovie(info, f)
				if err != nil {
					if !errors.Is(err, movie.NoResults) {
						return nil, err
					}
					m.Error = err.Error()
				}
			} else {
				m = &movie.Movie{
					Title:        info.Title,
					Year:         info.Year,
					Language:     info.Language,
					IsDir:        f.IsDir(),
					OriginalName: f.Name(),
				}
			}

			if err == nil {
				fmt.Printf("scan: %s => (%s, %d, %d)\n", f.Name(), m.Title, m.Year, m.ID)

				if m.ComputePath() {
					results.Results[m.OriginalName] = *m
				}
			} else {

				results.Failed[m.OriginalName] = *m
				fmt.Printf("scan: %s => (%v)\n", f.Name(), err)
			}
		}
	}

	results.Found = len(results.Results)
	results.Failures = len(results.Failed)
	log.Debugf("scanned files, found %d result(s)\n", results.Found)

	return results, nil
}
