package provider

import "github.com/TheoBrigitte/evansky/pkg/parser"

// Request represents a search request to a provider.
type Request struct {
	Query    string
	Language string
	Year     int

	// TODO: Maybe best to rely on presence of underlying struct to determine the request type?
	MediaType MediaType

	Movie *RequestMovie
	TV    *RequestTV
}

type RequestMovie struct {
	ID int
}

type RequestTV struct {
	ID      int
	Seasons []RequestTVSeason
}

type RequestTVSeason struct {
	ID       int
	Episodes []RequestTVEpisode
}

type RequestTVEpisode struct {
	ID int
}

// TODO: merge previous response and parsed information to create a more accurate request. Requests should be more accuracte as we walk down the directory tree.
// For example, if we have a parent directory named "Breaking Bad (2008)" and a child file named "S01E01 - Pilot.mkv", we should be able to create a request for the TV show "Breaking Bad" with season 1 and episode 1.
// TODO: handle movie collections: multiple movies in the same folder
func NewRequest(info parser.Info, resp Response) Request {
	return Request{
		Query: info.Title,
		Year:  info.Year,
	}
}
