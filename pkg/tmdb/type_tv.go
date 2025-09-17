package tmdb

import (
	"time"

	gotmdb "github.com/cyruzin/golang-tmdb"

	"github.com/TheoBrigitte/evansky/pkg/provider"
)

type tvResponse struct {
	gotmdb.TVShowResult

	firstAirDate time.Time
	mediaType    provider.MediaType
}

func newTVResponse(result gotmdb.TVShowResult) (*tvResponse, error) {
	// Parse the first air date in the format "2006-01-02"
	firstAirDate, err := time.Parse(time.DateOnly, result.FirstAirDate)
	if err != nil {
		return nil, err
	}

	t := &tvResponse{
		TVShowResult: result,
		firstAirDate: firstAirDate,
		mediaType:    provider.MediaTypeTV,
	}
	return t, nil
	//		response: response{
	//			ID:               t.ID,
	//			Name:             t.Name,
	//			OriginalName:     t.OriginalName,
	//			OriginalLanguage: t.OriginalLanguage,
	//			FirstAirDate:     t.FirstAirDate,
	//			VoteCount:        t.VoteCount,
	//			VoteAverage:      t.VoteAverage,
	//			//Popularity:       t.Popularity,
	//			MediaType: provider.MediaTypeTV,
	//		},
	//	}
}

func (r tvResponse) GetID() int {
	return int(r.ID)
}

func (r tvResponse) GetName() string {
	return r.Name
}

func (r tvResponse) GetDate() time.Time {
	return r.firstAirDate
}

func (r tvResponse) GetMediaType() provider.MediaType {
	return r.mediaType
}

func (r tvResponse) GetPopularity() int {
	return computePopularity(r.Popularity, r.VoteAverage, r.VoteCount)
}
