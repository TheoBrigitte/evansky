package tmdb

import (
	"time"

	gotmdb "github.com/cyruzin/golang-tmdb"

	"github.com/TheoBrigitte/evansky/pkg/provider"
)

type movieResponse struct {
	gotmdb.MovieResult

	releaseDate time.Time
	mediaType   provider.MediaType
}

func newMovieResponse(result gotmdb.MovieResult) (*movieResponse, error) {
	// Parse the release date in the format "2006-01-02"
	releaseDate, err := time.Parse(time.DateOnly, result.ReleaseDate)
	if err != nil {
		return nil, err
	}

	m := &movieResponse{
		MovieResult: result,
		releaseDate: releaseDate,
		mediaType:   provider.MediaTypeMovie,
	}

	return m, nil
	//	response: response{
	//		ID:               m.ID,
	//		Title:            m.Title,
	//		OriginalTitle:    m.OriginalTitle,
	//		OriginalLanguage: m.OriginalLanguage,
	//		ReleaseDate:      m.ReleaseDate,
	//		VoteCount:        m.VoteCount,
	//		VoteAverage:      m.VoteAverage,
	//		//Popularity:       m.Popularity,
	//		MediaType: provider.MediaTypeMovie,
	//	},
	//}
}

func (r movieResponse) GetID() int {
	return int(r.ID)
}

func (r movieResponse) GetName() string {
	return r.Title
}

func (r movieResponse) GetDate() time.Time {
	return r.releaseDate
}

func (r movieResponse) GetMediaType() provider.MediaType {
	return r.mediaType
}

func (r movieResponse) GetPopularity() int {
	return computePopularity(r.Popularity, r.VoteAverage, r.VoteCount)
}
