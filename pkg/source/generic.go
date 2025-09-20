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

type generic struct {
	path    string
	newPath string

	providers []provider.Interface

	nodes []Node

	// TODO: add setting to prefer file name preference over parent directories when finding a match
	// TODO: add settings to ignore n levels of directories (min-depth), and allow for max depth
}

func newGeneric(path string, providers []provider.Interface) (*generic, error) {
	s := &generic{
		path:      path,
		providers: providers,
	}

	return s, nil
}

func (g *generic) Process() error {
	info, err := os.Lstat(g.path)
	if err != nil {
		return err
	}
	dirInfo := fs.FileInfoToDirEntry(info)

	nodes, err := g.walk(g.path, dirInfo, nil)
	if err != nil {
		return err
	}
	g.nodes = append(g.nodes, nodes...)
	return nil
	//return filepath.WalkDir(g.path, g.startWalk)
}

// TODO: enrich parentReq with more info (like if it's a tv show or movie), then merge with current info as we walk down
func (g *generic) walk(path string, entry fs.DirEntry, parentResp provider.Response) ([]Node, error) {
	// Parse current file or directory name.
	info, err := parser.Parse(entry.Name())
	if err != nil {
		return nil, err
	}

	// TODO: if we have more information than the parent request, backtrack and re-query the providers with the new information. Year is most important.

	req := provider.Request{
		Query: info.Title,
		Year:  info.Year,

		Response: parentResp,
		Info:     *info,
		Entry:    entry,
	}

	var dirs []os.DirEntry
	if entry.IsDir() {
		dirs, err = os.ReadDir(path)
		if err != nil {
			return nil, err
		}
	}

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

	name := fmt.Sprintf("%s (%d)", resp.GetName(), resp.GetDate().Year())
	if !entry.IsDir() {
		// It's a file, create a single file source.
		dir := filepath.Dir(path)
		n := Node{
			PathOld: path,
			PathNew: filepath.Join(dir, name),
		}
		//slog.Info("found", "old", n.PathOld, "new", n.PathNew)
		return []Node{n}, nil
	}
	//slog.Info("found", "old", path, "new", name)
	req.Language = childLang
	resp.SetRequest(req)

	// TODO: Try to identify directory pattern (tv show, movie collection, etc).
	// Backtrack if we detect a different media type than the one we are looking for.

	var nodes []Node
	for _, nextEntry := range dirs {
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

//func (g *generic) startWalk(path string, d fs.DirEntry, err error) error {
//	if err != nil {
//		return err
//	}
//
//	// Parse current file or directory name.
//	info, err := parser.Parse(d.Name())
//	if err != nil {
//		return err
//	}
//
//	// Query the providers with the parsed information.
//	req := provider.Request{
//		Query:   info.Title,
//		Year:    info.Year,
//		Season:  info.Season,
//		Episode: info.Episode,
//	}
//	resp, err := g.Find(req, g.mediaType)
//	if err != nil {
//		slog.Error("processing", "path", path, "type", g.mediaType, "parsed", info, "error", err)
//		return nil
//	}
//	slog.Info("processing", "path", path, "type", g.mediaType, "parsed", info, "response", resp)
//
//	if !d.IsDir() {
//		// It's a file, create a single file source.
//		node := g.walkSingleFile(path, resp)
//		g.nodes = append(g.nodes, *node)
//		return nil
//	}
//
//	return fs.SkipDir
//}
//
//func (g *generic) walkSingleFile(path string, resp provider.Response) Source {
//	s := &generic{
//		path:      path,
//		newPath:   resp.GetPath(),
//		mediaType: resp.GetMediaType(),
//	}
//
//	return s
//}

// Find queries the providers in order until one returns a valid response.
// lang is the ISO 639-3 language code to use for the query.
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

func (g *generic) find(p provider.Interface, req provider.Request) (provider.Response, error) {
	if req.Response == nil {
		if req.Info.Title == "" {
			// We need at least a title to search.
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
		// Return Movie response as is.
		// TODO: detect if the new info is more accurate than the previous one, and re-query if needed.
		// we need the previous info to do that.
		return r, nil
	case provider.ResponseTV:
		// if isDir -> search for season then episode
		// else -> search for episode then season
		return g.findTVChild(p, r, req)
	case provider.ResponseTVSeason:
		if req.Info.Episode > 0 {
			// Get episode by number
			return r.GetEpisode(req.Info.Episode, req)
		}
		if req.Info.Title != "" {
			// Find episode by name.
			//req = g.usePreviousLanguage(req)
			return g.findTVEpisode(p, []provider.ResponseTVSeason{r}, req)
		}
		// We need episode information to go further.
		return nil, fmt.Errorf("find: no episode information")
	case provider.ResponseTVEpisode:
		// Return Episode response as is.
		// TODO: detect if the new info is more accurate than the previous one, and re-query if needed.
		// we need the previous info to do that.
		//return r, nil
		//newReq := g.usePreviousLanguage(req)
		newReq := req
		newReq.Response = r.GetSeason()
		return g.Find(newReq)
	}

	return nil, fmt.Errorf("find: unsupported response media type: %T", req.Response)
}

func (g *generic) searchByPopularity(p provider.Interface, req provider.Request) (provider.Response, error) {
	slog.Debug("searching by popularity", "query", req.Query, "year", req.Year)
	movie, err := p.SearchMovie(req)
	if err != nil && !errors.Is(err, provider.ErrNoResult) {
		return nil, err
	}

	tvshow, err := p.SearchTV(req)
	if err != nil && !errors.Is(err, provider.ErrNoResult) {
		return nil, err
	}

	if movie == nil && tvshow == nil {
		return nil, provider.ErrNoResult
	}

	if movie == nil {
		return tvshow, nil
	}

	if tvshow == nil {
		return movie, nil
	}

	if movie.GetPopularity() >= tvshow.GetPopularity() {
		return movie, nil
	}

	return tvshow, nil
}

func (g *generic) findTVChild(p provider.Interface, tv provider.ResponseTV, req provider.Request) (provider.Response, error) {
	if req.Info.Season > 0 {
		//req = g.usePreviousLanguage(req)

		// Get season by number
		season, err := tv.GetSeason(req.Info.Season, req)
		if err != nil {
			return nil, err
		}

		if req.Info.Episode > 0 {
			// Get episode by number
			return season.GetEpisode(req.Info.Episode, req)
		}

		return season, nil
	}
	if req.Info.Episode > 0 {
		//req = g.usePreviousLanguage(req)
		// Search for episode by number
		seasons, err := tv.GetSeasons(req)
		if err != nil {
			return nil, err
		}
		return g.findTVEpisode(p, seasons, req)
	}

	if req.Info.Title != "" {
		if req.Entry.IsDir() {
			// Try to detect season number from directory name
			seasonNumber, err := g.detectSeasonNumber(req.Info.Title)
			if err != nil {
				slog.Warn("findTVChild: cannot detect season number", "title", req.Info.Title, "error", err)
			}

			if seasonNumber > 0 {
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

func (g *generic) findTVSeasonOrEpisode(p provider.Interface, seasons []provider.ResponseTVSeason, req provider.Request) (provider.Response, error) {
	slog.Debug("find season or episode by name", "seasons", len(seasons), "title", req.Info.Title)
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

func (g *generic) findTVEpisode(p provider.Interface, seasons []provider.ResponseTVSeason, req provider.Request) (provider.Response, error) {
	slog.Debug("find episode", "seasons", len(seasons), "episode", req.Info.Episode, "title", req.Info.Title)
	if req.Info.Episode > 0 {
		for _, season := range seasons {
			return season.GetEpisode(req.Info.Episode, req)
		}

		return nil, fmt.Errorf("findTVEpisode: episode %d not found", req.Info.Episode)
	}

	if req.Info.Title == "" {
		return nil, fmt.Errorf("findTVEpisode: no episode information")
	}

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
