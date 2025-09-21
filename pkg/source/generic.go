package source

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"github.com/TheoBrigitte/evansky/pkg/parser"
	"github.com/TheoBrigitte/evansky/pkg/provider"
	"github.com/TheoBrigitte/evansky/pkg/source/language"
)

// generic implements a generic source that can handle both movies and TV shows.
// It provides functionality to scan directory structures, parse media information,
// and query metadata providers to generate nodes for renaming operations.
type generic struct {
	path     string   // Root path to scan for media files
	newPath  string   // Proposed new path (currently unused)
	options  Options  // Configuration options for scanning behavior
	excludes []string // List of files or directories to exclude based on glob patterns

	providers []provider.Interface // List of metadata providers to query
}

// newGeneric creates a new generic source instance with the specified path,
// providers, and options. The generic source can handle both movies and TV shows
// by querying metadata providers based on parsed file information.
func newGeneric(path string, providers []provider.Interface, o Options) *generic {
	s := &generic{
		path:      path,
		providers: providers,
	}

	return s
}

// scan scans the source path and returns a list of nodes.
// It starts by getting information about the root path and then recursively
// walks the directory tree to process each file and directory.
func (g *generic) scan() ([]Node, error) {
	// Get initial file or directory info.
	info, err := os.Lstat(g.path)
	if err != nil {
		return nil, err
	}
	dirInfo := fs.FileInfoToDirEntry(info)

	if g.options.ExcludeGlob != "" {
		excludes, err := filepath.Glob(filepath.Join(g.path, g.options.ExcludeGlob))
		if err != nil {
			return nil, err
		}
		g.excludes = excludes
	}

	// Start walking the directory tree.
	nodes, err := g.walk(g.path, dirInfo, nil)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

// walk recursively walks the directory tree and processes each file or directory.
// For each entry, it:
// 1. Parses the name to extract media information
// 2. Detects the language based on directory contents
// 3. Queries metadata providers to get accurate information
// 4. Generates nodes for files or continues recursion for directories
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
	} else {
		shouldExclude := slices.Contains(g.excludes, path)
		if shouldExclude {
			slog.Info("excluding", "path", path)
			// TODO: this might need better handling to avoid nil pointer dereference error
			return nil, nil
		}
	}

	// Detect the language of the media based on multiple factors.
	// confidence is only used for logging purposes.
	lang, confidence, childLang := language.Detect(req, dirs)
	req.Language = lang

	//slog.Debug("processing", "info", info, "request", req, "path", path, "confidence", lang.Confidence, "reliable", lang.IsReliable())
	slog.Info("searching", "path", path, "language", req.Language, "confidence", confidence, "parent", parentResp != nil)

	// Query the providers with the parsed information.
	resp, err := g.Find(req)
	if err != nil {
		n := Node{
			Entry: entry,
			Error: err,
			Path:  path,
			Info:  *info,
		}
		//slog.Info("found", "old", n.PathOld, "new", n.PathNew)

		slog.Error("processed", "error", err, "path", path)
		return []Node{n}, nil
	}
	slog.Info("found", "name", resp.GetName(), "year", resp.GetDate().Year(), "type", fmt.Sprintf("%T", resp))
	//slog.Debug("processed", "response", resp, "path", path)

	if !entry.IsDir() {
		// This is a file, generate a rename node for it.
		//name := fmt.Sprintf("%s (%d)", resp.GetName(), resp.GetDate().Year())
		//dir := filepath.Dir(path)

		n := Node{
			Entry:    entry,
			Path:     path,
			Info:     *info,
			Response: resp,
		}
		//slog.Info("found", "old", n.PathOld, "new", n.PathNew)

		return []Node{n}, nil
	}
	//slog.Info("found", "old", path, "new", name)

	// This is a directory, continue walking.
	// Enforce the detected language for child entries, as this is more accurate since
	// language was detected over all child entries.
	req.Language = childLang
	resp.SetRequest(req)

	// TODO: Try to identify directory pattern (tv show, movie collection, etc).
	// Backtrack if we detect a different media type than the one we are looking for.

	var nodes []Node
	for _, nextEntry := range dirs {
		// Build the next path as: current path + entry name.
		nextPath := filepath.Join(path, nextEntry.Name())
		// TODO: allow for non-recursive scan
		childNodes, err := g.walk(nextPath, nextEntry, resp)
		if err != nil {
			return nil, err
		}
		if childNodes == nil {
			continue
		}
		nodes = append(nodes, childNodes...)
	}

	return nodes, nil
}

// Find queries all providers in order until one returns a valid response.
// It tries each provider sequentially and returns the first successful result.
// If all providers fail, it returns an error.
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
// It makes a decisions based on the request information and previous response:
// - For top-level media:
//   - If season or episode information is provided, searches for TV shows
//   - Otherwise searches for movies or TV shows by popularity
//
// - For Movie: returns the movie response directly
// - For TV show: searches for seasons or episodes based on available information
// - For movies: returns the movie response directly
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
// This method is used when we have ambiguous media that could be either a movie or TV show.
// It queries both endpoints and compares popularity scores to determine the best match.
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

// usePreviousLanguage copies the language setting from a previous request to maintain
// language consistency throughout the processing tree. This ensures that once a language
// is detected at a higher level, it's propagated to child requests.
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
