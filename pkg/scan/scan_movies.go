package scan

import (
	"errors"
	"os"
	"strconv"

	parse "github.com/middelink/go-parse-torrent-name"
	log "github.com/sirupsen/logrus"

	"github.com/TheoBrigitte/evansky/pkg/movie"
)

func (s *Scanner) searchMovie(info *parse.TorrentInfo, f os.FileInfo) (*movie.Movie, error) {
	search := &movie.Movie{
		Title:        info.Title,
		Year:         info.Year,
		Language:     info.Language,
		IsDir:        f.IsDir(),
		OriginalName: f.Name(),
	}

	// Search movie with title and year.
	log.Debugf("search title=%s year=%s\n", search.Title, strconv.Itoa(search.Year))
	movies, err := s.client.GetMovies(search.Title, strconv.Itoa(search.Year))
	if err != nil {
		return nil, err
	}

	result, err := movie.Best(movies)
	if errors.Is(err, movie.NoResults) {
		// Search movie with title only, and match year after.
		log.Debugf("search hit NoResults error, trying without exact year.\n")
		movies, err = s.client.GetMovies(search.Title, "")
		if err != nil {
			return nil, err
		}

		result, err = movie.BestByYear(movies, search.Year)
	}
	if err != nil {
		return nil, err
	}

	search.ID = result.ID
	search.Title = result.Title
	search.Year = result.Year
	search.Language = result.Language

	return search, nil
}
