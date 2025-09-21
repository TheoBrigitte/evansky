package source

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/TheoBrigitte/evansky/pkg/provider"
)

// findTVChild finds a TV show child (season or episode) based on the request information.
// It handles different scenarios:
// - Season number provided: gets the specific season, optionally with episode
// - Episode number only: searches across all seasons for the episode
// - Title only: attempts season number detection or searches by name
func (g *generic) findTVChild(p provider.Interface, tv provider.ResponseTV, req provider.Request) (provider.Response, error) {
	resp, err := g.findTVChildWithNumber(p, tv, req)
	if err != nil {
		return nil, err
	}
	if resp != nil {
		return resp, nil
	}

	if req.Query != "" {
		// Only title provided, try to detect season number from directory name if possible

		if req.Entry.IsDir() {
			// Try to detect season number from directory name
			seasonNumber, err := extractNumber(req.Query, seasonRegex)
			if err != nil {
				slog.Warn("findTVChild: cannot detect season number", "title", req.Query, "error", err)
			}

		} else {
			// Try to detect episode number from directory name
			episodeNumber, err := extractNumber(req.Query, episodeRegex)
			if err != nil {
				slog.Warn("findTVChild: cannot detect episode number", "title", req.Query, "error", err)
			}
		}
		if seasonNumber > 0 || episodeNumber > 0 {
			return g.findTVChildWithNumber(p, tv, req)
		}

		// Search for season or episode by name
		return g.findTVSeasonOrEpisode(p, tv.GetSeasons(), req)
	}

	return nil, fmt.Errorf("findTVChild: no season or episode information")
}

func (g *generic) findTVChildWithNumber(p provider.Interface, tv provider.ResponseTV, req provider.Request) (provider.Response, error) {
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

	return nil, nil
}

// findTVSeasonOrEpisode finds a TV show season or episode based on the request information.
// It finds the best match among all seasons and episodes, by
// comparing the request title against season names and episode names.
func (g *generic) findTVSeasonOrEpisode(p provider.Interface, seasons []provider.ResponseTVSeason, req provider.Request) (provider.Response, error) {
	slog.Debug("find season or episode by name", "seasons", len(seasons), "title", req.Query, "season", req.Info.Season, "episode", req.Info.Episode)

	// Search for season or episode by name using.
	var bestMatch provider.Response
	var bestScore float64 = -1
	//seasons := make([]gotmdb.TVSeason, 0, len(show.Seasons))
	for _, season := range seasons {
		isBetter, seasonScore := betterMatch(req.Query, season.GetName(), bestScore)
		if isBetter {
			bestScore = seasonScore
			bestMatch = season
		}

		for _, episode := range season.GetEpisodes() {
			isBetter, episodeScore := betterMatch(req.Query, episode.GetName(), bestScore)
			if isBetter {
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
// - Episode title provided: find the best matching episode name
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

	// Search for episode by finding the best match.
	var bestMatch provider.Response
	var bestScore float64 = -1
	for _, season := range seasons {
		for _, episode := range season.GetEpisodes() {
			isBetter, episodeScore := betterMatch(req.Query, episode.GetName(), bestScore)
			if isBetter {
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
