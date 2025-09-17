package source

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/TheoBrigitte/evansky/pkg/parser"
	"github.com/TheoBrigitte/evansky/pkg/provider"
)

type generic struct {
	path      string
	newPath   string
	mediaType provider.MediaType

	providers []provider.Interface

	nodes []Node

	// TODO: add setting to prefer file name preference over parent directories when finding a match
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

	// Query the providers with the parsed information.
	req := provider.NewRequest(*info, parentResp)
	resp, err := g.Find(req, g.mediaType)
	if err != nil {
		slog.Debug("processing", "path", path, "type", g.mediaType, "parsed", info, "error", err)
		return nil, nil
	}
	slog.Debug("processing", "path", path, "type", g.mediaType, "parsed", info, "response", resp)

	if !entry.IsDir() {
		// It's a file, create a single file source.
		dir := filepath.Dir(path)
		name := fmt.Sprintf("%s (%d)", resp.GetName(), resp.GetDate().Year())
		n := Node{
			PathOld: path,
			PathNew: filepath.Join(dir, name),
		}
		slog.Info("found", "old", n.PathOld, "new", n.PathNew)
		return []Node{n}, nil
	}

	dirs, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	// TODO: Try to identify directory pattern (tv show, movie collection, etc).
	// Backtrack if we detect a different media type than the one we are looking for.

	var nodes []Node
	for _, nextEntry := range dirs {
		nextPath := filepath.Join(path, nextEntry.Name())
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

func (g *generic) Find(req provider.Request, mediaType provider.MediaType) (provider.Response, error) {
	for _, p := range g.providers {
		responses, err := p.Search(req)
		if err != nil {
			slog.Debug("provider search error", "provider", p.Name(), "error", err)
			continue
		}

		return responses[0], nil
	}

	return nil, fmt.Errorf("no result")
}
