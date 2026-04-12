package source

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/TheoBrigitte/evansky/pkg/parser"
	"github.com/TheoBrigitte/evansky/pkg/provider"
	"github.com/TheoBrigitte/evansky/pkg/source/language"
)

var (
	ErrExcludedPath = errors.New("excluded path")
	ErrRetryable    = errors.New("retryable error")
)

// generic implements a generic source that can handle both movies and TV shows.
// It provides functionality to scan directory structures, parse media information,
// and query metadata providers to generate nodes for renaming operations.
type generic struct {
	path         string         // Root path to scan for media files
	options      Options        // Configuration options for scanning behavior
	excludes     []string       // List of files or directories to exclude based on glob patterns
	excludeRegex *regexp.Regexp // Compiled regex for excluding files or directories
	includeRegex *regexp.Regexp // Compiled regex for include files or directories
	titleRegex   *regexp.Regexp // Compiled regex for extracting title from file or directory name

	providers []provider.Interface // List of metadata providers to query
}

// scan scans the source path and returns a list of nodes.
// It starts by getting information about the root path and then recursively
// walks the directory tree to process each file and directory.
func (g *generic) scan() ([]Node, error) {
	// Get initial file or directory info.
	info, err := os.Lstat(g.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read path info %s: %w", g.path, err)
	}
	dirInfo := fs.FileInfoToDirEntry(info)

	if g.options.ExcludeGlob != "" {
		// filepaht.Glob requires a full path to match, so we need to join the exclude glob with the base path.
		excludePath := g.path
		if !dirInfo.IsDir() {
			// In case of a file, we need to use the parent directory as the base path for the glob pattern.
			excludePath = filepath.Dir(g.path)
		}
		excludes, err := filepath.Glob(filepath.Join(excludePath, g.options.ExcludeGlob))
		if err != nil {
			return nil, fmt.Errorf("failed to apply exclude pattern: %w", err)
		}
		g.excludes = excludes
	}

	if g.options.ExcludeRegex != "" {
		g.excludeRegex, err = regexp.Compile(g.options.ExcludeRegex)
		if err != nil {
			return nil, fmt.Errorf("failed to compile exclude regex: %w", err)
		}
	}

	if g.options.IncludeRegex != "" {
		g.includeRegex, err = regexp.Compile(g.options.IncludeRegex)
		if err != nil {
			return nil, fmt.Errorf("failed to compile include regex: %w", err)
		}
	}

	if g.options.TitleRegex != "" {
		g.titleRegex, err = regexp.Compile(g.options.TitleRegex)
		if err != nil {
			return nil, fmt.Errorf("failed to compile title regex: %w", err)
		}
	}

	// Start walking the directory tree.
	return g.walk(g.path, dirInfo, 0, nil), nil
}

// walk recursively walks the directory tree and processes each file or directory.
// For each entry, it:
// 1. Parses the name to extract media information
// 2. Detects the language based on directory contents
// 3. Queries metadata providers to get accurate information
// 4. Generates nodes for files or continues recursion for directories
func (g *generic) walk(path string, entry fs.DirEntry, depth int, parentResp provider.Response) []Node {
	n := Node{
		Entry: entry,
		Path:  path,
	}

	// Set node type based on file extension.
	extension := strings.TrimPrefix(strings.ToLower(filepath.Ext(entry.Name())), ".")
	switch {
	case slices.Contains(g.options.MediaExts, extension):
		n.Type = NodeTypeMedia
	case slices.Contains(g.options.SubtitleExts, extension):
		n.Type = NodeTypeSubtitle
	}

	// Skip entries that are explicitly excluded by glob patterns.
	if slices.Contains(g.excludes, path) {
		n.Error = fmt.Errorf("%w by glob", ErrExcludedPath)
		return []Node{n}
	}

	// Skip entries that are excluded by regex patterns.
	if g.excludeRegex != nil && g.excludeRegex.MatchString(path) {
		n.Error = fmt.Errorf("%w by regex", ErrExcludedPath)
		return []Node{n}
	}

	// Skip entries that are not included by regex patterns.
	// Only check non-directory entries for include regex, as we want traverse directories and find only matching files.
	if g.includeRegex != nil && !entry.IsDir() && !g.includeRegex.MatchString(path) {
		n.Error = fmt.Errorf("%w not included by regex", ErrExcludedPath)
		return []Node{n}
	}

	log.Debug().Str("path", path).Msgf("scanning")

	// query default to file or directory name
	query := entry.Name()
	if g.options.Query != "" && depth == 0 {
		// use user provided query override if set
		query = g.options.Query
	}
	if g.titleRegex != nil {
		matches := g.titleRegex.FindStringSubmatch(query)
		if len(matches) > 1 {
			query = matches[len(matches)-1]
		}
	}

	// Parse current query to extract media information.
	info, err := parser.Parse(query)
	if err != nil {
		n.Error = fmt.Errorf("failed to query %s: %w", query, err)
		return []Node{n}
	}

	n.Info = *info

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

	log.Debug().Str("path", path).Str("query", req.Query).Int("Year", info.Year).Interface("info", info).Msg("parsed media info")

	if info.Year == 0 && info.Season == 0 && info.Episode == 0 {
		// Parsing did not yield useful information, use the full name as query.
		// This fixes an issue where some names are not parsed correctly.
		req.Query = filepath.Base(parser.CleanTitle(req.Query))
		req.Query = strings.TrimSuffix(req.Query, filepath.Ext(req.Query))
	}

	// Read directory entries early if it's a directory.
	// This is needed for language detection.
	var dirs []os.DirEntry
	if entry.IsDir() {
		dirs, err = os.ReadDir(path)
		if err != nil {
			n.Error = fmt.Errorf("failed to read directory %s: %w", path, err)
			return []Node{n}
		}
	}

	// Detect the language of the media based on multiple factors.
	// confidence is only used for logging purposes.
	lang, confidence, childLang := language.Detect(req, dirs)
	req.QueryLanguage = lang

	// Override language if query language is set.
	if g.options.QueryLanguage != "" {
		req.QueryLanguage = g.options.QueryLanguage
	}

	req.DestinationLanguage = g.options.Language

	// slog.Debug("processing", "info", info, "request", req, "path", path, "confidence", lang.Confidence, "reliable", lang.IsReliable())
	log.Debug().Str("language", req.QueryLanguage).Float64("confidence", confidence).Msgf("detected language")

	var resp provider.Response
	if g.options.StripComponents <= depth {
		// Query the providers with the parsed information.
		resp, err = g.Find(req)
		if err != nil {
			// slog.Info("found", "old", n.PathOld, "new", n.PathNew)
			n.Error = fmt.Errorf("failed to find media: %w", err)

			// log.Err(err).Str("path", path).Msg("processed")
			return []Node{n}
		}

		log.Debug().Int("id", resp.GetID()).Str("name", resp.GetName()).Int("year", resp.GetDate().Year()).Str("type", fmt.Sprintf("%T", resp)).Msgf("found    %s", path)

		n.Response = resp
		// This is a directory, continue walking.
		// Enforce the detected language for child entries, as this is more accurate since
		// language was detected over all child entries.
		req.QueryLanguage = childLang
		resp.SetRequest(req)
	} else {
		log.Debug().Msgf("skipping entry due to strip components setting (depth %d, strip %d): %s", depth, g.options.StripComponents, path)
	}
	depth++

	// slog.Debug("processed", "response", resp, "path", path)

	if !entry.IsDir() {
		// This is a file, generate a rename node for it.
		// name := fmt.Sprintf("%s (%d)", resp.GetName(), resp.GetDate().Year())
		// dir := filepath.Dir(path)

		// slog.Info("found", "old", n.PathOld, "new", n.PathNew)

		return []Node{n}
	}
	// slog.Info("found", "old", path, "new", name)

	// TODO: Try to identify directory pattern (tv show, movie collection, etc).
	// Backtrack if we detect a different media type than the one we are looking for.

	var nodes []Node
	for _, nextEntry := range dirs {
		// Build the next path as: current path + entry name.
		nextPath := filepath.Join(path, nextEntry.Name())
		// TODO: allow for non-recursive scan
		childNodes := g.walk(nextPath, nextEntry, depth, resp)
		if childNodes == nil {
			continue
		}
		nodes = append(nodes, childNodes...)
	}

	return nodes
}

// Find queries all providers in order until one returns a valid response.
// It tries each provider sequentially and returns the first successful result.
// If all providers fail, it returns an error.
func (g *generic) Find(req provider.Request) (provider.Response, error) {
	for _, p := range g.providers {
		resp, err := g.find(p, req)
		if err != nil {
			log.Debug().Err(err).Str("provider", p.Name()).Msg("provider search failed")
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("no result")
}

// find queries a single provider with the given request and returns a response.
// It makes a decisions based on the request information and previous response:
// - For top-level media:
//   - If season or episode information is provided, searches for TV shows
//   - Otherwise searches for movies or TV shows by popularity
//
// - For Movie: returns the movie response directly
// - For TV show: searches for seasons or episodes based on available information
// - For movies: returns the movie response directly
func (g *generic) find(p provider.Interface, req provider.Request) (provider.Response, error) {
	log.Debug().Str("query", req.Query).Int("year", req.Year).Str("language", req.QueryLanguage).Int("season", req.Info.Season).Int("episode", req.Info.Episode).Str("response", fmt.Sprintf("%T", req.Response)).Msgf("finding media")
	if req.Response == nil {
		// Processing a top level media (no previous response).

		if req.Query == "" {
			// We need at least query to search for top level media.
			return nil, fmt.Errorf("find: no query")
		}

		if req.Info.Season > 0 || req.Info.Episode > 0 {
			// Search for TV show season or episode.
			tv, _, err := p.SearchTV(req)
			if err != nil {
				return nil, err
			}
			return g.findTVChild(p, tv, req)
		}

		// Search for Movie or TV show.
		return g.searchByYearOrPopularity(p, req)
	}

	// Change language of the response to match the request language.
	resp, err := req.Response.InLanguage(req)
	if err != nil {
		return nil, err
	}

	switch r := resp.(type) {
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
			return r.GetEpisode(req.Info.Episode)
		}
		if req.Query != "" {
			// Parent is a season, search for episode by name.
			// req = g.usePreviousLanguage(req)
			return g.findTVEpisode(p, []provider.ResponseTVSeason{r}, req)
		}
		// We need episode information to go further.
		return nil, fmt.Errorf("find: no episode information")
	case provider.ResponseTVEpisode:
		// Parent is an episode, handle this as a sibling episode.
		// TODO: detect if the new info is more accurate than the previous one, and re-query if needed.
		// we need the previous info to do that.
		// newReq := g.usePreviousLanguage(req)
		newReq := req
		newReq.Response = r.GetSeason()
		return g.Find(newReq)
	}

	return nil, fmt.Errorf("find: unsupported response media type: %T", req.Response)
}

// searchByYearOrPopularity searches for both movie and TV show and returns the best match.
// This method is used when we have ambiguous media that could be either a movie or TV show.
// It queries both endpoints and compares using a combined score that considers both
// title similarity, year proximity, and popularity to determine the best match.
func (g *generic) searchByYearOrPopularity(p provider.Interface, req provider.Request) (provider.Response, error) {
	movie, movieScore, err := p.SearchMovie(req)
	if err != nil && !errors.Is(err, provider.ErrNoResult) {
		// Ignore no result error, as we want to try TV show search as well.
		return nil, err
	}

	tvshow, tvScore, err := p.SearchTV(req)
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

	// Both movie and TV show found, use their scores to decide which is better.
	// Scores are calculated by the search functions and include title similarity,
	// year proximity, and popularity.

	log.Debug().
		Str("movie_name", movie.GetName()).
		Int("movie_year", movie.GetDate().Year()).
		Int("movie_popularity", movie.GetPopularity()).
		Float64("movie_score", movieScore).
		Str("tv_name", tvshow.GetName()).
		Int("tv_year", tvshow.GetDate().Year()).
		Int("tv_popularity", tvshow.GetPopularity()).
		Float64("tv_score", tvScore).
		Msg("comparing movie and tv show by combined score")

	if movieScore <= tvScore {
		// Movie has better (lower) combined score.
		return movie, nil
	}

	// TV show has better (lower) combined score.
	return tvshow, nil
}
