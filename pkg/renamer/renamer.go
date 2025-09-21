package renamer

import (
	"fmt"
	"log/slog"
	"maps"
	"path/filepath"
	"slices"

	"github.com/TheoBrigitte/evansky/pkg/provider"
	"github.com/TheoBrigitte/evansky/pkg/renamer/format"
	"github.com/TheoBrigitte/evansky/pkg/source"
)

var (
	deduplicationAttemptLimit = 2
	deduplicationSuffix       = "_"
)

type Renamer interface {
	Run(dryRun, force bool) error
}

type renamer struct {
	// List of directories to create
	directories map[string]struct{}

	// List of files to create (source -> destination)
	files      map[string]string
	filesError map[string]error

	o         Options
	paths     []string
	providers []provider.Interface
}

type Options struct {
	DryRun     bool
	Formatter  format.Formatter
	RenameMode string
	Output     string
}

func New(paths []string, providers []provider.Interface, o Options) (Renamer, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("at least one source is required")
	}
	if len(providers) == 0 {
		return nil, fmt.Errorf("at least one provider is required")
	}

	r := &renamer{
		directories: make(map[string]struct{}),
		files:       make(map[string]string),
		filesError:  make(map[string]error),
		o:           o,
		paths:       paths,
		providers:   providers,
	}

	return r, nil
}

func (r *renamer) Run(dryRun, force bool) error {
	prefix := ""
	if dryRun {
		prefix = "[dry-run] "
	}

	o := source.Options{}

	var nodes []source.Node
	for _, path := range r.paths {
		n, err := source.Scan(path, r.providers, o)
		if err != nil {
			slog.Error("scan failed", "path", path, "error", err)
			continue
		}
		nodes = append(nodes, n...)
	}

	if len(nodes) == 0 {
		slog.Warn("no nodes found")
		return nil
	}

	r.processNodes(nodes)

	for dir := range maps.Keys(r.directories) {
		fmt.Printf("%s[new directory] -> [%s]\n", prefix, dir)
	}

	for src, dst := range r.files {
		fmt.Printf("%s[%s] -> [%s]\n", prefix, src, dst)
	}

	for src, err := range r.filesError {
		fmt.Printf("%s[%s] -> error: %s\n", prefix, src, err)
	}

	// TODO: handle non destructive renaming, keeping other files (subtitles, etc)
	// TODO: include directories in the renaming process

	return nil
}

func (r *renamer) processNodes(nodes []source.Node) {
	//w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	//fmt.Fprintln(w, "Path\tName\tYear\tError\tType")

	for _, n := range nodes {
		if n.Error != nil {
			r.filesError[n.Path] = n.Error
			continue
		}

		newPath, err := r.generateName(n)
		if err != nil {
			r.filesError[n.Path] = err
			continue
		}
		r.files[n.Path] = newPath

		//fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%T\n", n.Path, name, year, errMsg, n.Response)
	}

	//w.Flush()
}

func (r *renamer) generateName(node source.Node) (string, error) {
	components := []string{}
	switch resp := node.Response.(type) {
	case provider.ResponseMovie:
		components = r.o.Formatter.Movie(resp)
	case provider.ResponseTV:
		components = r.o.Formatter.TVShow(resp)
	case provider.ResponseTVSeason:
		components = r.o.Formatter.TVSeason(resp)
	case provider.ResponseTVEpisode:
		components = r.o.Formatter.TVEpisode(resp)
	default:
		return "", fmt.Errorf("unknown type: %T", node.Response)
	}
	if len(components) == 0 {
		return "", fmt.Errorf("no components")
	}

	var dir string
	newPath := filepath.Join(components...)
	if len(components) > 1 {
		dir = filepath.Dir(newPath)
		r.directories[dir] = struct{}{}
	}

	if _, exists := r.files[node.Path]; exists {
		return "", fmt.Errorf("duplicate source path")
	}

	for i := 0; i <= deduplicationAttemptLimit; i++ {
		exists := slices.Contains(slices.Collect(maps.Values(r.files)), newPath)
		if !exists {
			return newPath, nil
		}

		newPath = fmt.Sprintf("%s%s", newPath, deduplicationSuffix)
	}

	return "", fmt.Errorf("could not deduplicate path")
}
