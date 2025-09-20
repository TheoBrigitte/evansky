package source

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/gogf/gf/v2/text/gstr"

	"github.com/TheoBrigitte/evansky/pkg/parser"
	"github.com/TheoBrigitte/evansky/pkg/provider"
	"github.com/TheoBrigitte/evansky/pkg/source/language"
)

// generic implements a generic source that can handle both movies and TV shows.
type generic struct {
	path    string
	newPath string
	options Options

	providers []provider.Interface
}

// newGeneric creates a new generic source.
func newGeneric(path string, providers []provider.Interface, o Options) *generic {
	s := &generic{
		path:      path,
		providers: providers,
	}

	return s
}

// scan scans the source path and returns a list of nodes to rename.
func (g *generic) scan() ([]Node, error) {
	// Get initial file or directory info.
	info, err := os.Lstat(g.path)
	if err != nil {
		return nil, err
	}
	dirInfo := fs.FileInfoToDirEntry(info)

	// Start walking the directory tree.
	nodes, err := g.walk(g.path, dirInfo, nil)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

// walk recursively walks the directory tree and processes each file or directory.
func (g *generic) walk(path string, entry fs.DirEntry, parentResp provider.Response) ([]Node, error) {
	// Parse current file or directory name to extract media information.
	info, err := parser.Parse(entry.Name())
	if err != nil {
		return nil, err
	}

	// TODO: if we have more information than the parent request, backtrack and re-query the providers with the new information. Year is most important.

	// Create a new request with the parsed information and the parent response.
	req := provider.Request{
		// Parsed information
		Query: info.Title,
		Year:  info.Year,

		// File system information
		Info:  *info,
		Entry: entry,

		// Parent response
		Response: parentResp,
	}

	// Read directory entries early if it's a directory.
	// This is needed for language detection.
	var dirs []os.DirEntry
	if entry.IsDir() {
		dirs, err = os.ReadDir(path)
		if err != nil {
			return nil, err
		}
	}

	// Detect the language of the media based multiple factors.
	// confidence is only used for logging purposes.
	lang, confidence, childLang := language.Detect(req, dirs)
	req.Language = lang

	//slog.Debug("processing", "info", info, "request", req, "path", path, "confidence", lang.Confidence, "reliable", lang.IsReliable())
	slog.Info("searching", "path", path, "language", req.Language, "confidence", confidence, "parent", parentResp != nil)

	// Query the providers with the parsed information.
	resp, err := g.Find(req)
	if err != nil {
		slog.Error("processed", "error", err, "path", path)
		return nil, nil
	}
	slog.Info("found", "name", resp.GetName(), "year", resp.GetDate().Year(), "type", fmt.Sprintf("%T", resp))
	//slog.Debug("processed", "response", resp, "path", path)

	if !entry.IsDir() {
		// This is a file, generate a node for it.

		name := fmt.Sprintf("%s (%d)", resp.GetName(), resp.GetDate().Year())
		dir := filepath.Dir(path)

		n := Node{
			PathOld: path,
			PathNew: filepath.Join(dir, name),
		}
		//slog.Info("found", "old", n.PathOld, "new", n.PathNew)

		return []Node{n}, nil
	}
	//slog.Info("found", "old", path, "new", name)

	// This is a directory, continue walking.
	// Enforce the detected language for child entries, as this is more accurate since
	// language was detect over all child entries.
	req.Language = childLang
	resp.SetRequest(req)

	// TODO: Try to identify directory pattern (tv show, movie collection, etc).
	// Backtrack if we detect a different media type than the one we are looking for.

	var nodes []Node
	for _, nextEntry := range dirs {
		// Build the next path as: current path + entry name.
		nextPath := filepath.Join(path, nextEntry.Name())
		// TODO: allow for non-recursive scan
		nodes, err := g.walk(nextPath, nextEntry, resp)
		if err != nil {
			return nil, err
		}
		if nodes == nil {
			continue
		}
		nodes = append(nodes, nodes...)
	}

	return nodes, nil
}

// Find queries all providers in order until one returns a valid response.
func (g *generic) Find(req provider.Request) (provider.Response, error) {
	for _, p := range g.providers {
		resp, err := g.find(p, req)
		if err != nil {
			slog.Warn("provider search error", "provider", p.Name(), "error", err)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("no result")
}

// find queries a single provider with the given request and returns a response.
// it makes decisions based on the request information and previous response (if any).
func (g *generic) find(p provider.Interface, req provider.Request) (provider.Response, error) {
	if req.Response == nil {
		// Processing a top level media (no previous response).

		if req.Info.Title == "" {
			// We need at least a title to search for top level media.
			return nil, fmt.Errorf("find: no title")
		}

		if req.Info.Season > 0 || req.Info.Episode > 0 {
			// Search for TV show season or episode.
			tv, err := p.SearchTV(req)
			if err != nil {
				return nil, err
			}
			return g.findTVChild(p, tv, req)
		}

		// Search for Movie or TV show.
		return g.searchByPopularity(p, req)
	}

	switch r := req.Response.(type) {
	case provider.ResponseMovie:
		// Parent response media is a movie, return it as is.
		// TODO: go back if we detect a TV show from the new info.
		// TODO: detect if the new info is more accurate than the previous one, and re-query if needed.
		// we need the previous info to do that.
		return r, nil
	case provider.ResponseTV:
		// Parent is a TV show, search for season or episode.
		// TODO: implement child search order
		// if isDir -> search for season then episode
		// else -> search for episode then season
		return g.findTVChild(p, r, req)
	case provider.ResponseTVSeason:
		if req.Info.Episode > 0 {
			// Parent is a season, get the episode by number.
			return r.GetEpisode(req.Info.Episode, req)
		}
		if req.Info.Title != "" {
			// Parent is a season, search for episode by name.
			//req = g.usePreviousLanguage(req)
			return g.findTVEpisode(p, []provider.ResponseTVSeason{r}, req)
		}
		// We need episode information to go further.
		return nil, fmt.Errorf("find: no episode information")
	case provider.ResponseTVEpisode:
		// Parent is an episode, handle this as a sibling episode.
		// TODO: detect if the new info is more accurate than the previous one, and re-query if needed.
		// we need the previous info to do that.
		//newReq := g.usePreviousLanguage(req)
		newReq := req
		newReq.Response = r.GetSeason()
		return g.Find(newReq)
	}

	return nil, fmt.Errorf("find: unsupported response media type: %T", req.Response)
}

// searchByPopularity searches for both movie and TV show and returns the most popular one.
func (g *generic) searchByPopularity(p provider.Interface, req provider.Request) (provider.Response, error) {
	slog.Debug("searching by popularity", "query", req.Query, "year", req.Year)
	movie, err := p.SearchMovie(req)
	if err != nil && !errors.Is(err, provider.ErrNoResult) {
		// Ignore no result error, as we want to try TV show search as well.
		return nil, err
	}

	tvshow, err := p.SearchTV(req)
	if err != nil && !errors.Is(err, provider.ErrNoResult) {
		// Ignore no result error, as we want to try movie search as well.
		return nil, err
	}

	if movie == nil && tvshow == nil {
		// No result from either search.
		return nil, provider.ErrNoResult
	}

	if movie == nil {
		// Only TV show found.
		return tvshow, nil
	}

	if tvshow == nil {
		// Only movie found.
		return movie, nil
	}

	if movie.GetPopularity() >= tvshow.GetPopularity() {
		// Movie is more popular than TV show.
		return movie, nil
	}

	// TV show is more popular than movie.
	return tvshow, nil
}

// findTVChild finds a TV show child (season or episode) based on the request information.
func (g *generic) findTVChild(p provider.Interface, tv provider.ResponseTV, req provider.Request) (provider.Response, error) {
	if req.Info.Season > 0 {
		// Prefer season number if available

		//req = g.usePreviousLanguage(req)

		// Get season by number
		season, err := tv.GetSeason(req.Info.Season, req)
		if err != nil {
			return nil, err
		}

		if req.Info.Episode > 0 {
			// Season and episode number provided, get the episode
			return season.GetEpisode(req.Info.Episode, req)
		}

		// Only season number provided, return the season
		return season, nil
	}

	if req.Info.Episode > 0 {
		// Only episode number provided, search for episode by name
		//req = g.usePreviousLanguage(req)
		seasons, err := tv.GetSeasons(req)
		if err != nil {
			return nil, err
		}
		return g.findTVEpisode(p, seasons, req)
	}

	if req.Info.Title != "" {
		// Only title provided, try to detect season number from directory name if possible

		if req.Entry.IsDir() {
			// Try to detect season number from directory name
			seasonNumber, err := g.detectSeasonNumber(req.Info.Title)
			if err != nil {
				slog.Warn("findTVChild: cannot detect season number", "title", req.Info.Title, "error", err)
			}

			if seasonNumber > 0 {
				// Season number detected, get the season
				return tv.GetSeason(seasonNumber, req)
			}
		}

		// Search for season or episode by name
		seasons, err := tv.GetSeasons(req)
		if err != nil {
			return nil, err
		}
		return g.findTVSeasonOrEpisode(p, seasons, req)
	}

	return nil, fmt.Errorf("findTVChild: no season or episode information")
}

// findTVSeasonOrEpisode finds a TV show season or episode based on the request information.
func (g *generic) findTVSeasonOrEpisode(p provider.Interface, seasons []provider.ResponseTVSeason, req provider.Request) (provider.Response, error) {
	slog.Debug("find season or episode by name", "seasons", len(seasons), "title", req.Info.Title)

	// Search for season or episode by name using Levenshtein distance to find the best match.
	var bestMatch provider.Response
	var bestScore int = -1
	//seasons := make([]gotmdb.TVSeason, 0, len(show.Seasons))
	for _, season := range seasons {
		seasonScore := gstr.Levenshtein(req.Info.Title, season.GetName(), 1, 1, 1)
		if bestScore == -1 || seasonScore < bestScore {
			bestScore = seasonScore
			bestMatch = season
		}

		episodes, err := season.GetEpisodes(req)
		if err != nil {
			slog.Warn("findTVSeasonOrEpisode: cannot get episodes", "showID", season.GetShow().GetID(), "season", season.GetSeasonNumber(), "error", err)
			continue
		}
		for _, episode := range episodes {
			episodeScore := gstr.Levenshtein(req.Info.Title, episode.GetName(), 1, 1, 1)
			if bestScore == -1 || episodeScore < bestScore {
				bestScore = episodeScore
				bestMatch = episode
			}
		}
	}

	if bestMatch != nil {
		return bestMatch, nil
	}

	return nil, fmt.Errorf("findTVSeasonOrEpisode: no match found for %s", req.Info.Title)
}

// findTVEpisode finds a TV show episode based on the request information.
func (g *generic) findTVEpisode(p provider.Interface, seasons []provider.ResponseTVSeason, req provider.Request) (provider.Response, error) {
	slog.Debug("find episode", "seasons", len(seasons), "episode", req.Info.Episode, "title", req.Info.Title)
	if req.Info.Episode > 0 {
		// Episode number provided, get the episode from the first season that has it.
		for _, season := range seasons {
			return season.GetEpisode(req.Info.Episode, req)
		}

		return nil, fmt.Errorf("findTVEpisode: episode %d not found", req.Info.Episode)
	}

	if req.Info.Title == "" {
		// We need at least an episode title to search for an episode.
		return nil, fmt.Errorf("findTVEpisode: no episode information")
	}

	// Search for episode by name using Levenshtein distance to find the best match.
	var bestMatch provider.Response
	var bestScore int = -1
	for _, season := range seasons {
		episodes, err := season.GetEpisodes(req)
		if err != nil {
			slog.Warn("findTVEpisode: cannot get episodes", "showID", season.GetShow().GetID(), "season", season.GetSeasonNumber(), "error", err)
			continue
		}

		for _, episode := range episodes {
			episodeScore := gstr.Levenshtein(req.Info.Title, episode.GetName(), 1, 1, 1)
			if bestScore == -1 || episodeScore < bestScore {
				bestScore = episodeScore
				bestMatch = episode
			}
		}
	}

	if bestMatch != nil {
		return bestMatch, nil
	}

	return nil, fmt.Errorf("findTVEpisode: episode %s no match found", req.Info.Title)
}

var seasonRegex = regexp.MustCompile(`[0-9]+`)

// detectSeasonNumber tries to detect a season number from a string.
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

func (g *generic) usePreviousLanguage(req provider.Request) provider.Request {
	if req.Response == nil {
		return req
	}

	prevReq := req.Response.GetRequest()
	if prevReq == nil || prevReq.Language == "" {
		return req
	}

	req.Language = prevReq.Language
	return req
}
