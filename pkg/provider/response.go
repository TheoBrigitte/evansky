package provider

import "time"

// Response represents a search response from a provider.
type Response interface {
	// Commont methods for all media types.
	GetID() int
	GetName() string
	GetDate() time.Time
	GetPopularity() int

	ResponseBase
}

type ResponseMovie interface {
	Response

	ResponseBaseMovie
}

type ResponseTV interface {
	Response

	GetSeason(int, Request) (ResponseTVSeason, error)
	GetSeasons(Request) ([]ResponseTVSeason, error)

	ResponseBaseTV
}

type ResponseTVSeason interface {
	Response

	GetShow() ResponseTV
	GetSeasonNumber() int
	GetEpisode(int, Request) (ResponseTVEpisode, error)
	GetEpisodes(Request) ([]ResponseTVEpisode, error)

	ResponseBaseTVSeason
}

type ResponseTVEpisode interface {
	Response

	GetEpisodeNumber() int
	GetSeason() ResponseTVSeason

	ResponseBaseTVEpisode
}

//GetMediaType() MediaType

type ResponseBase interface {
	GetRequest() *Request
	SetRequest(Request)
}

type responseBase struct {
	Request *Request
}

func newResponseBase() *responseBase {
	return &responseBase{}
}

func (r *responseBase) GetRequest() *Request {
	return r.Request
}

func (r *responseBase) SetRequest(req Request) {
	r.Request = &req
}

type ResponseBaseMovie interface {
	ResponseBase
	movie()
}

type responseBaseMovie struct {
	ResponseBase
}

func (r responseBaseMovie) movie() {}

func NewResponseBaseMovie() *responseBaseMovie {
	return &responseBaseMovie{
		ResponseBase: newResponseBase(),
	}
}

type ResponseBaseTV interface {
	ResponseBase
	tv()
}

type responseBaseTV struct {
	ResponseBase
}

func (r responseBaseTV) tv() {}

func NewResponseBaseTV() *responseBaseTV {
	return &responseBaseTV{
		ResponseBase: newResponseBase(),
	}
}

type ResponseBaseTVSeason interface {
	ResponseBase
	tvSeason()
}

type responseBaseTVSeason struct {
	ResponseBase
}

func (r responseBaseTVSeason) tvSeason() {}

func NewResponseBaseTVSeason() *responseBaseTVSeason {
	return &responseBaseTVSeason{
		ResponseBase: newResponseBase(),
	}
}

type ResponseBaseTVEpisode interface {
	ResponseBase
	tvEpisode()
}

type responseBaseTVEpisode struct {
	ResponseBase
}

func (r responseBaseTVEpisode) tvEpisode() {}

func NewResponseBaseTVEpisode() *responseBaseTVEpisode {
	return &responseBaseTVEpisode{
		ResponseBase: newResponseBase(),
	}
}
