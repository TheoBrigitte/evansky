package scan

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	parse "github.com/middelink/go-parse-torrent-name"
	log "github.com/sirupsen/logrus"

	"github.com/TheoBrigitte/evansky/pkg/movie"
)

func (s *Scanner) searchMulti(info *parse.TorrentInfo, f os.FileInfo) (*movie.Multi, error) {
	log.Debugf("searching info: %#v\n", info)
	search := &movie.Multi{
		Title:        info.Title,
		Year:         info.Year,
		Language:     info.Language,
		IsDir:        f.IsDir(),
		OriginalName: f.Name(),
	}

	// Search multi with title and year.
	log.Debugf("search title=%s year=%s\n", search.Title, strconv.Itoa(search.Year))
	multi, err := s.client.GetMulti(search.Title, strconv.Itoa(search.Year))
	if err != nil {
		return nil, err
	}

	result, err := movie.BestMulti(multi)
	if errors.Is(err, movie.NoResults) {
		// Search movie with title only, and match year after.
		log.Debugf("search hit NoResults error, trying without exact year.\n")
		multi, err = s.client.GetMulti(search.Title, "")
		if err != nil {
			return nil, err
		}

		result, err = movie.BestByYearMulti(multi, search.Year)
	}
	log.Debugf("search result: %#v\n", result)

	if result != nil {
		switch result.MediaType {
		case "movie":
			// nothing to do
		default:
			return nil, fmt.Errorf("%w: %s", movie.UnsupportedMediaType, result.MediaType)
		}

		search.ID = result.ID
		search.Title = result.Title
		search.Year = result.Year
		search.Language = result.Language
		search.MediaType = result.MediaType
	}

	return search, nil
}
