package scan

import (
	"errors"
	"fmt"
	"os"
	"strconv"

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

			m := &movie.Movie{
				Title:        info.Title,
				Year:         info.Year,
				Language:     info.Language,
				IsDir:        f.IsDir(),
				OriginalName: f.Name(),
			}

			if !s.noAPI {
				var m2 *movie.Movie
				m2, err = s.searchMovie(*m)
				if err != nil {
					if !errors.Is(err, movie.NoResults) {
						return nil, err
					}
					m.Error = err.Error()
				} else {
					m.ID = m2.ID
					m.Title = m2.Title
					m.Year = m2.Year
					m.Language = m2.Language
				}
			}

			if err == nil {
				fmt.Printf("scan: %s => (%s, %d, %d)\n", f.Name(), m.Title, m.Year, m.ID)

				if m.ComputePath() {
					results.Results[m.OriginalName] = *m
				}
			}
		}
	}

	results.Found = len(results.Results)
	log.Debugf("scanned files, found %d result(s)\n", results.Found)

	return results, nil
}

func (s *Scanner) searchMovie(search movie.Movie) (*movie.Movie, error) {
	// Search movie with title and year.
	log.Debugf("search title=%s year=%s\n", search.Title, strconv.Itoa(search.Year))
	movies, err := s.client.GetMovies(search.Title, strconv.Itoa(search.Year))
	if err != nil {
		return nil, err
	}

	m1, err := movie.Best(movies)
	if errors.Is(err, movie.NoResults) {
		// Search movie with title only, and match year after.
		log.Debugf("search hit NoResults error, trying without exact year.\n")
		movies, err = s.client.GetMovies(search.Title, "")
		if err != nil {
			return nil, err
		}

		m2, err := movie.BestByYear(movies, search.Year)
		if err != nil {
			return nil, err
		}

		return m2, nil
	}
	if err != nil {
		return nil, err
	}

	return m1, nil
}
