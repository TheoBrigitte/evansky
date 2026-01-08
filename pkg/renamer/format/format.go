package format

import (
	"github.com/TheoBrigitte/evansky/pkg/provider"
	"github.com/TheoBrigitte/evansky/pkg/source"
)

type Formatter interface {
	Movie(provider.ResponseMovie, source.Node) []string
	TVShow(provider.ResponseTV, source.Node) []string
	TVSeason(provider.ResponseTVSeason, source.Node) []string
	TVEpisode(provider.ResponseTVEpisode, source.Node) []string
	FileSuffix(string, source.Node) string
}
