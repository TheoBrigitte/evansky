package provider

import (
	"errors"
	"fmt"

	"github.com/TheoBrigitte/evansky/pkg/parser"
)

// Request represents a search request to a provider.
type Request struct {
	Query    string
	Year     int
	Language string

	// TODO: Maybe best to rely on presence of underlying struct to determine the request type?
	MediaType MediaType

	Movie *RequestMovie
	TV    *RequestTV
}

type RequestMovie struct {
	ID int
}

type RequestTV struct {
	ID            int
	SeasonNumber  int
	EpisodeNumber int
}

// TODO: merge previous response and parsed information to create a more accurate request. Requests should be more accuracte as we walk down the directory tree.
// For example, if we have a parent directory named "Breaking Bad (2008)" and a child file named "S01E01 - Pilot.mkv", we should be able to create a request for the TV show "Breaking Bad" with season 1 and episode 1.
func NewRequest(info parser.Info, req *Request, resp Response) (*Request, error) {
	r := &Request{
		Query: info.Title,
		Year:  info.Year,
		TV:       &RequestTV{},
	}

	if info.Season > 0 || info.Episode > 0 {
		// If we have season or episode information, it's a TV show.
		r.MediaType = MediaTypeTV
		r.TV.SeasonNumber = info.Season
		r.TV.EpisodeNumber = info.Episode
	}

	if resp == nil || req == nil {
		// Case 1: first request
		return r, nil
	}

	if req.MediaType == MediaTypeUnknown {
		// Case 2: second request, previous request was unknown
		// TODO: maybe ignore if current info holds tv season/episode
		if info.Year > 0 && req.Year != info.Year {
			// If we have different year than previous request, use the current one and ignore result.
			// e.g. parent directory is "Movie (2020)" and child file is "Movie (2021).mkv"
			return r, nil
		}
	}

	reqType := req.MediaType
	respType := resp.GetMediaType()
	switch respType {
	case MediaTypeMovie:
		if !oneOf(reqType, MediaTypeMovie, MediaTypeUnknown) {
			// If previous request was not a movie or unknown, we have a conflict.
			// TODO: implement collection
			return nil, conflictError(respType, reqType)
		}
		r.MediaType = MediaTypeMovie
		r.Movie = &RequestMovie{
			ID: resp.GetID(),
		}
		return r, nil
	case MediaTypeTV:
		if oneOf(reqType, MediaTypeMovie) {
			return nil, conflictError(respType, reqType)
		}
		r.MediaType = MediaTypeTV
		r.TV.ID = resp.GetShowID()
		return r, nil
	case MediaTypeTVSeason:
		if oneOf(reqType, MediaTypeMovie) {
			return nil, conflictError(respType, reqType)
		}
		r.MediaType = MediaTypeTV
		r.TV.ID = resp.GetShowID()
		r.TV.SeasonNumber = resp.GetSeasonNumber()
		return r, nil
	case MediaTypeTVEpisode:
		if oneOf(reqType, MediaTypeMovie) {
			return nil, conflictError(respType, reqType)
		}
		r.MediaType = MediaTypeTV
		r.TV.ID = resp.GetShowID()
		r.TV.SeasonNumber = resp.GetSeasonNumber()
		r.TV.EpisodeNumber = resp.GetEpisodeNumber()
		return r, nil
	}

	return r, nil
}

func (r Request) String() string {
	return fmt.Sprintf("Request{Query: %q, Year: %d, Language: %q MediaType: %s, Movie: %+v, TV: %+v}", r.Query, r.Year, r.Language, r.MediaType, r.Movie, r.TV)
}

func oneOf(a MediaType, others ...MediaType) bool {
	for _, o := range others {
		if a == o {
			return true
		}
	}
	return false
}

var MediaTypeConflictError = errors.New("conflicting media types")

func conflictError(a, b MediaType) error {
	return fmt.Errorf("%w: response is a %s, but previous request is %s", MediaTypeConflictError, a, b)
}
