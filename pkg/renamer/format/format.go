package format

import (
	"github.com/TheoBrigitte/evansky/pkg/provider"
)

type Formatter interface {
	Movie(provider.ResponseMovie) []string
	TVShow(provider.ResponseTV) []string
	TVSeason(provider.ResponseTVSeason) []string
	TVEpisode(provider.ResponseTVEpisode) []string
}
