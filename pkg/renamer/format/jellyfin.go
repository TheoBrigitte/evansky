package format

import (
	"fmt"
	"strconv"

	"github.com/TheoBrigitte/evansky/pkg/provider"
	"github.com/TheoBrigitte/evansky/pkg/source"
	"github.com/TheoBrigitte/evansky/pkg/source/language"
)

type JellyfinFormatter struct{}

func NewJellyfinFormatter() JellyfinFormatter {
	return JellyfinFormatter{}
}

// Movie format according to Jellyfin's recommended naming conventions.
// https://jellyfin.org/docs/general/server/media/movies
func (f JellyfinFormatter) Movie(m provider.ResponseMovie, n source.Node) []string {
	movieFormat := fmt.Sprintf("%s (%d)", m.GetName(), m.GetDate().Year())
	movieFormat = f.setSubtitleLanguage(movieFormat, n)
	return []string{movieFormat, movieFormat}
}

// TVShow format according to Jellyfin's recommended naming conventions.
// https://jellyfin.org/docs/general/server/media/shows
func (f JellyfinFormatter) TVShow(tv provider.ResponseTV, n source.Node) []string {
	return []string{fmt.Sprintf("%s (%d)", tv.GetName(), tv.GetDate().Year())}
}

// TVSeason format according to Jellyfin's recommended naming conventions.
func (f JellyfinFormatter) TVSeason(s provider.ResponseTVSeason, n source.Node) []string {
	showFormat := f.TVShow(s.GetShow(), n)

	seasonPadding := len(strconv.Itoa(len(s.GetShow().GetSeasons())))
	seasonFormat := fmt.Sprintf("Season %0*d", seasonPadding, s.GetSeasonNumber())

	return append(showFormat, seasonFormat)
}

// TVEpisode format according to Jellyfin's recommended naming conventions.
func (f JellyfinFormatter) TVEpisode(e provider.ResponseTVEpisode, n source.Node) []string {
	seasonFormat := f.TVSeason(e.GetSeason(), n)

	season := e.GetSeason()
	show := season.GetShow()

	// Padding based on total number of seasons/episodes
	// e.g. S01E01 for less than 10 seasons/episodes, S001E001 for less than 100 seasons/episodes, etc.
	seasonPadding := len(strconv.Itoa(len(show.GetSeasons())))
	episodePadding := len(strconv.Itoa(len(season.GetEpisodes())))

	episodeFormat := fmt.Sprintf("%s - S%0*dE%0*d - %s", show.GetName(), seasonPadding, season.GetSeasonNumber(), episodePadding, e.GetEpisodeNumber(), e.GetName())
	episodeFormat = f.setSubtitleLanguage(episodeFormat, n)

	return append(seasonFormat, episodeFormat)
}

func (f JellyfinFormatter) setSubtitleLanguage(name string, n source.Node) string {
	if n.Type == source.NodeTypeSubtitle {
		normalizedLang := language.NormalizeLanguage(n.Info.Language)
		if normalizedLang != "" {
			return fmt.Sprintf("%s.%s", name, normalizedLang)
		}
	}

	return name
}
