package source

import (
	"fmt"
	"log/slog"
	"regexp"
	"strconv"

	"github.com/gogf/gf/v2/text/gstr"

	"github.com/TheoBrigitte/evansky/pkg/provider"
)

// findTVChild finds a TV show child (season or episode) based on the request information.
// It handles different scenarios:
// - Season number provided: gets the specific season, optionally with episode
// - Episode number only: searches across all seasons for the episode
// - Title only: attempts season number detection or searches by name
func (g *generic) findTVChild(p provider.Interface, tv provider.ResponseTV, req provider.Request) (provider.Response, error) {
	if req.Info.Season > 0 {
		// Prefer season number if available

		//req = g.usePreviousLanguage(req)

		// Get season by number
		season, err := tv.GetSeason(req.Info.Season)
		if err != nil {
			return nil, err
		}

		if req.Info.Episode > 0 {
			// Season and episode number provided, get the episode
			return season.GetEpisode(req.Info.Episode)
		}

		// Only season number provided, return the season
		return season, nil
	}

	if req.Info.Episode > 0 {
		// Only episode number provided, search for episode across all seasons
		//req = g.usePreviousLanguage(req)
		return g.findTVEpisode(p, tv.GetSeasons(), req)
	}

	if req.Query != "" {
		// Only title provided, try to detect season number from directory name if possible

		if req.Entry.IsDir() {
			// Try to detect season number from directory name
			seasonNumber, err := g.detectSeasonNumber(req.Query)
			if err != nil {
				slog.Warn("findTVChild: cannot detect season number", "title", req.Query, "error", err)
			}

			if seasonNumber > 0 {
				// Season number detected, get the season
				return tv.GetSeason(seasonNumber)
			}
		}

		// Search for season or episode by name
		return g.findTVSeasonOrEpisode(p, tv.GetSeasons(), req)
	}

	return nil, fmt.Errorf("findTVChild: no season or episode information")
}

// findTVSeasonOrEpisode finds a TV show season or episode based on the request information.
// It uses Levenshtein distance to find the best match among all seasons and episodes,
// comparing the request title against season names and episode names.
func (g *generic) findTVSeasonOrEpisode(p provider.Interface, seasons []provider.ResponseTVSeason, req provider.Request) (provider.Response, error) {
	slog.Debug("find season or episode by name", "seasons", len(seasons), "title", req.Query, "season", req.Info.Season, "episode", req.Info.Episode)

	// Search for season or episode by name using Levenshtein distance to find the best match.
	var bestMatch provider.Response
	var bestScore int = -1
	//seasons := make([]gotmdb.TVSeason, 0, len(show.Seasons))
	for _, season := range seasons {
		seasonScore := gstr.Levenshtein(req.Query, season.GetName(), 1, 1, 1)
		if bestScore == -1 || seasonScore < bestScore {
			bestScore = seasonScore
			bestMatch = season
		}

		for _, episode := range season.GetEpisodes() {
			episodeScore := gstr.Levenshtein(req.Query, episode.GetName(), 1, 1, 1)
			if bestScore == -1 || episodeScore < bestScore {
				bestScore = episodeScore
				bestMatch = episode
			}
		}
	}

	if bestMatch != nil {
		return bestMatch, nil
	}

	return nil, fmt.Errorf("findTVSeasonOrEpisode: no match found for %s", req.Query)
}

// findTVEpisode finds a TV show episode based on the request information.
// It handles two scenarios:
// - Episode number provided: searches for the episode by number across seasons
// - Episode title provided: uses Levenshtein distance to find the best matching episode name
func (g *generic) findTVEpisode(p provider.Interface, seasons []provider.ResponseTVSeason, req provider.Request) (provider.Response, error) {
	slog.Debug("find episode", "seasons", len(seasons), "title", req.Query, "season", req.Info.Season, "episode", req.Info.Episode)
	if req.Info.Episode > 0 {
		// Episode number provided, get the episode from the first season that has it.
		for _, season := range seasons {
			return season.GetEpisode(req.Info.Episode)
		}

		return nil, fmt.Errorf("findTVEpisode: episode %d not found", req.Info.Episode)
	}

	if req.Query == "" {
		// We need at least an episode title to search for an episode.
		return nil, fmt.Errorf("findTVEpisode: no episode information")
	}

	// Search for episode by name using Levenshtein distance to find the best match.
	var bestMatch provider.Response
	var bestScore int = -1
	for _, season := range seasons {
		for _, episode := range season.GetEpisodes() {
			episodeScore := gstr.Levenshtein(req.Query, episode.GetName(), 1, 1, 1)
			if bestScore == -1 || episodeScore < bestScore {
				bestScore = episodeScore
				bestMatch = episode
			}
		}
	}

	if bestMatch != nil {
		return bestMatch, nil
	}

	return nil, fmt.Errorf("findTVEpisode: episode %s no match found", req.Query)
}

// seasonRegex is a compiled regular expression used to extract numeric values
// from season directory names for season number detection.
var seasonRegex = regexp.MustCompile(`[0-9]+`)

// detectSeasonNumber tries to detect a season number from a directory name string.
// It extracts all numeric sequences from the name and returns the last one as the season number.
// This heuristic works for common season directory naming patterns like "Season 1", "S02", etc.
func (g *generic) detectSeasonNumber(name string) (int, error) {
	matches := seasonRegex.FindAllString(name, -1)
	if len(matches) > 0 {
		// Convert the last match to an integer.
		seasonNumber, err := strconv.Atoi(matches[len(matches)-1])
		if err != nil {
			return -1, err
		}
		return seasonNumber, nil
	}

	return -1, nil
}
