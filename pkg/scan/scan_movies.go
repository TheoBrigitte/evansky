package scan

import (
	"errors"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/TheoBrigitte/evansky/pkg/movie"
)

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
