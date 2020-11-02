package scan

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"

	log "github.com/sirupsen/logrus"

	gotmdb "github.com/cyruzin/golang-tmdb"

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

			var title string
			var year int
			var id int64
			var language string
			if !s.noAPI {
				var movies *gotmdb.SearchMovies
				log.Debugf("search title=%s year=%s\n", info.Title, strconv.Itoa(info.Year))
				movies, err = s.client.GetMovies(info.Title, strconv.Itoa(info.Year))
				if err != nil {
					return nil, err
				}

				var m *movie.Movie
				m, err = movie.Best(movies)
				if errors.Is(err, movie.NoResults) {
					log.Debugf("search hit NoResults error, trying without exact year.\n")
					movies, err = s.client.GetMovies(info.Title, "")
					if err != nil {
						return nil, err
					}

					m, err = movie.BestByYear(movies, info.Year)
				}
				if err == nil {
					title = m.Title
					year = m.ReleaseDate.Year()
					language = m.Language
					id = m.ID
				}
			} else {
				title = info.Title
				year = info.Year
				language = info.Language
			}
			if err != nil {
				//return err
				fmt.Printf("scan: %s => (%v)\n", f.Name(), err)
			} else {
				fmt.Printf("scan: %s => (%s, %d, %d)\n", f.Name(), title, year, id)

				var newPath string
				if f.IsDir() {
					newPath = movie.Path(title, year)
				} else {
					newPath = path.Join(movie.Path(title, year), f.Name())
				}

				if newPath != f.Name() {
					results.Results[f.Name()] = Result{
						ID:       id,
						Language: language,
						Path:     newPath,
						IsDir:    f.IsDir(),
					}
				}
			}
		}
	}
	results.Found = len(results.Results)
	log.Debugf("scanned files, found %d result(s)\n", results.Found)

	return results, nil
}
