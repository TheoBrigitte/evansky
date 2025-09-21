package renamer

import (
	"fmt"
	"io"
	"log/slog"
	"maps"
	"os"
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
	Run(source.Options) error
}

type renamer struct {
	// List of directories to create
	directories map[string]struct{}

	// List of files to create (source -> destination)
	files  map[string]string
	errors map[string]error

	o         Options
	paths     []string
	providers []provider.Interface
}

type Options struct {
	Write      bool
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
		errors:      make(map[string]error),
		o:           o,
		paths:       paths,
		providers:   providers,
	}

	return r, nil
}

func (r *renamer) Run(o source.Options) (err error) {
	prefix := ""
	if !r.o.Write {
		prefix = "[dry-run] "
	}

	var w writer
	switch r.o.RenameMode {
	case "symlink":
		w = os.Symlink
	case "copy":
		w = copyFile
	default:
		return fmt.Errorf("unknown rename mode: %q", r.o.RenameMode)
	}

	var nodes = make(map[string][]source.Node)
	for _, path := range r.paths {
		output := r.o.Output
		if output == "" {
			output = filepath.Dir(filepath.Clean(path))
		}

		n, err := source.Scan(path, r.providers, o)
		if err != nil {
			slog.Error("scan failed", "path", path, "error", err)
			continue
		}
		nodes[output] = append(nodes[output], n...)
	}

	if len(nodes) == 0 {
		slog.Warn("no nodes found")
		return nil
	}

	r.processNodes(nodes)

	slog.Info("### starting renaming", "files", len(r.files), "directories", len(r.directories))

	for dir := range maps.Keys(r.directories) {
		if r.o.Write {
			err := os.MkdirAll(dir, 0750)
			if err != nil {
				return fmt.Errorf("failed to create directory %q: %w", dir, err)
			}
		}
		fmt.Printf("%s[new directory] -> [%s]\n", prefix, dir)
	}
	if len(r.directories) > 0 {
		slog.Info("### directories created", "count", len(r.directories))
	}

	for src, dst := range r.files {
		realSrc := src
		if r.o.RenameMode == "symlink" {
			realSrc, err = getSymlinkSrc(src, dst)
			if err != nil {
				r.errors[src] = err
				continue
			}
		}

		// Perform the write operation
		err := r.write(realSrc, dst, w)
		if err != nil {
			r.errors[src] = err
			continue
		}
		fmt.Printf("%s[%s] -> [%s]\n", prefix, src, dst)
	}
	if len(r.files) > 0 {
		slog.Info("### files renamed", "count", len(r.files))
	}

	for src, err := range r.errors {
		fmt.Printf("%s[%s] -> error: %s\n", prefix, src, err)
	}
	if len(r.errors) > 0 {
		slog.Warn("### errors encountered", "count", len(r.errors))
	}

	// TODO: handle non destructive renaming, keeping other files (subtitles, etc)
	// TODO: include directories in the renaming process

	return nil
}

func (r *renamer) processNodes(nodesOutput map[string][]source.Node) {
	//w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	//fmt.Fprintln(w, "Path\tName\tYear\tError\tType")

	for output, nodes := range nodesOutput {
		for _, n := range nodes {
			if n.Error != nil {
				r.errors[n.Path] = n.Error
				continue
			}

			newPath, err := r.generateName(n, output)
			if err != nil {
				r.errors[n.Path] = err
				continue
			}
			r.files[n.Path] = newPath

			//fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%T\n", n.Path, name, year, errMsg, n.Response)
		}
	}

	//w.Flush()
}

func (r *renamer) generateName(node source.Node, output string) (string, error) {
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

	var extension string
	if !node.Entry.IsDir() {
		extension = filepath.Ext(node.Entry.Name())
	}

	var dir string
	// Prepend output directory if specified
	newPath := filepath.Join(append([]string{output}, components...)...)
	newPathWithExt := filepath.Clean(fmt.Sprintf("%s%s", newPath, extension))
	if node.Path == newPathWithExt {
		return "", fmt.Errorf("source and destination are the same")
	}

	if len(components) > 1 {
		dir = filepath.Dir(newPathWithExt)
		r.directories[dir] = struct{}{}
	}

	if _, exists := r.files[node.Path]; exists {
		return "", fmt.Errorf("duplicate source path")
	}

	for i := 0; i <= deduplicationAttemptLimit; i++ {
		exists := slices.Contains(slices.Collect(maps.Values(r.files)), newPathWithExt)
		if !exists {
			return newPathWithExt, nil
		}

		newPath = fmt.Sprintf("%s%s", newPath, deduplicationSuffix)
		newPathWithExt = filepath.Clean(fmt.Sprintf("%s%s", newPath, extension))
	}

	return "", fmt.Errorf("could not deduplicate path")
}

type writer func(string, string) error

func (r *renamer) write(src, dst string, fn writer) error {
	if !r.o.Write || fn == nil {
		return nil
	}

	return fn(src, dst)
}

func getSymlinkSrc(src, dst string) (string, error) {
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return "", err
	}
	absDst, err := filepath.Abs(filepath.Dir(dst))
	if err != nil {
		return "", err
	}
	realSrc, err := filepath.Rel(absDst, absSrc)
	if err != nil {
		return "", err
	}

	return realSrc, nil
}

func copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}
