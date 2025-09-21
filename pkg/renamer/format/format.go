package format

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/TheoBrigitte/evansky/pkg/provider"
)

type Formatter interface {
	Movie(provider.ResponseMovie) string
	TVShow(provider.ResponseTV) string
	TVSeason(provider.ResponseTVSeason) string
	TVEpisode(provider.ResponseTVEpisode) string
}

type JellyfinFormatter struct {
}

func NewJellyfinFormatter() JellyfinFormatter {
	return JellyfinFormatter{}
}

// https://jellyfin.org/docs/general/server/media/movies
func (f JellyfinFormatter) Movie(m provider.ResponseMovie) string {
	return fmt.Sprintf("%s (%d)", m.GetName(), m.GetDate().Year())
}

// https://jellyfin.org/docs/general/server/media/shows
func (f JellyfinFormatter) TVShow(tv provider.ResponseTV) string {
	return fmt.Sprintf("%s (%d)", tv.GetName(), tv.GetDate().Year())
}

func (f JellyfinFormatter) TVSeason(s provider.ResponseTVSeason) string {
	showFormat := f.TVShow(s.GetShow())

	seasonPadding := len(strconv.Itoa(len(s.GetShow().GetSeasons())))
	seasonFormat := fmt.Sprintf("Season %0*d", seasonPadding, s.GetSeasonNumber())

	return filepath.Join(showFormat, seasonFormat)
}

func (f JellyfinFormatter) TVEpisode(e provider.ResponseTVEpisode) string {
	seasonFormat := f.TVSeason(e.GetSeason())

	season := e.GetSeason()
	show := season.GetShow()

	// Padding based on total number of seasons/episodes
	// e.g. S01E01 for less than 10 seasons/episodes, S001E001 for less than 100 seasons/episodes, etc.
	seasonPadding := len(strconv.Itoa(len(show.GetSeasons())))
	episodePadding := len(strconv.Itoa(len(season.GetEpisodes())))

	episodeFormat := fmt.Sprintf("%s S%0*dE%0*d", show.GetName(), seasonPadding, season.GetSeasonNumber(), episodePadding, e.GetEpisodeNumber())

	return filepath.Join(seasonFormat, episodeFormat)
}
