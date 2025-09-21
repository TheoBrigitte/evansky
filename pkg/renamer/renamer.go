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

type Entry struct {
	Source      string
	Destination string
	Error       error
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

	entries := []Entry{}
	dirs := []string{}
	for output, nodes := range nodes {
		for _, n := range nodes {
			entry, dir := r.generateEntry(n, output)
			entries = append(entries, entry)
			if dir != "" {
				dirs = append(dirs, dir)
			}
		}
	}
	slices.Sort(dirs)

	slog.Info("### starting renaming", "entries", len(entries))

	dirCount := 0
	for _, dir := range slices.Compact(dirs) {
		if r.o.Write {
			err := os.MkdirAll(dir, 0750)
			if err != nil {
				return fmt.Errorf("failed to create directory %q: %w", dir, err)
			}
		}
		fmt.Printf("%s[new directory] -> [%s]\n", prefix, dir)
		dirCount++
	}
	if dirCount > 0 {
		slog.Info("### directories created", "count", dirCount)
	}

	filesCount := 0
	for _, e := range entries {
		if e.Error != nil {
			continue
		}

		realSrc := e.Source
		if r.o.RenameMode == "symlink" {
			realSrc, err = getSymlinkSrc(e.Source, e.Destination)
			if err != nil {
				e.Error = err
				continue
			}
		}

		// Perform the write operation
		err := r.write(realSrc, e.Destination, w)
		if err != nil {
			e.Error = err
			continue
		}

		fmt.Printf("%s[%s] -> [%s]\n", prefix, e.Source, e.Destination)
		filesCount++
	}
	if filesCount > 0 {
		slog.Info("### files renamed", "count", filesCount)
	}

	errorsCount := 0
	for _, e := range entries {
		if e.Error == nil {
			continue
		}

		fmt.Printf("%s[%s] -> error: %s\n", prefix, e.Source, e.Error)
		errorsCount++
	}
	if errorsCount > 0 {
		slog.Warn("### errors encountered", "count", errorsCount)
	}

	// TODO: handle non destructive renaming, keeping other files (subtitles, etc)
	// TODO: include directories in the renaming process

	return nil
}

func (r *renamer) generateEntry(node source.Node, output string) (e Entry, dir string) {
	e.Source = node.Path

	if node.Error != nil {
		e.Error = node.Error
		return
	}

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
		e.Error = fmt.Errorf("unknown type: %T", node.Response)
		return
	}
	if len(components) == 0 {
		e.Error = fmt.Errorf("no components")
		return
	}

	var extension string
	if !node.Entry.IsDir() {
		extension = filepath.Ext(node.Entry.Name())
	}

	// Prepend output directory if specified
	newPath := filepath.Join(append([]string{output}, components...)...)
	newPathWithExt := filepath.Clean(fmt.Sprintf("%s%s", newPath, extension))
	if node.Path == newPathWithExt {
		e.Error = fmt.Errorf("source and destination are the same")
		return
	}

	if len(components) > 1 {
		// Add directories to the list of directories to create
		dir = filepath.Dir(newPathWithExt)
	}

	if _, exists := r.files[node.Path]; exists {
		e.Error = fmt.Errorf("duplicate source path")
		return
	}

	for i := 0; i <= deduplicationAttemptLimit; i++ {
		exists := slices.Contains(slices.Collect(maps.Values(r.files)), newPathWithExt)
		if !exists {
			e.Destination = newPathWithExt
			return
		}

		newPath = fmt.Sprintf("%s%s", newPath, deduplicationSuffix)
		newPathWithExt = filepath.Clean(fmt.Sprintf("%s%s", newPath, extension))
	}

	e.Error = fmt.Errorf("could not deduplicate path")
	return
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
