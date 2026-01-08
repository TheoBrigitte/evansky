package renamer

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/rs/zerolog/log"

	"github.com/TheoBrigitte/evansky/pkg/provider"
	"github.com/TheoBrigitte/evansky/pkg/renamer/format"
	"github.com/TheoBrigitte/evansky/pkg/source"
)

var (
	// deduplicationAttemptLimit defines the maximum number of attempts to deduplicate
	// destination paths by appending suffixes
	deduplicationAttemptLimit = 2
	// deduplicationSuffix is the string appended to filenames when resolving path conflicts
	deduplicationSuffix = "_"
)

// Renamer defines the interface for renaming media files based on provider metadata.
type Renamer interface {
	// Run executes the renaming process with the given source options
	Run(source.Options) error
}

// renamer implements the Renamer interface and manages the state of the renaming process.
type renamer struct {
	// directories tracks directories that need to be created
	directories map[string]struct{}

	// files maps source paths to their destination paths
	files map[string]string
	// errors tracks errors encountered during entry generation
	errors map[string]error

	// o contains the configuration options for the renamer
	o Options
	// paths contains the source paths to scan for media files
	paths []string
	// providers contains the metadata providers to use for lookups
	providers []provider.Interface
}

// Options configures the behavior of the renamer.
type Options struct {
	// Force enables overwriting existing files at the destination
	Force bool
	// Formatter defines how to format the destination filenames
	Formatter format.Formatter
	// Output specifies the base directory for renamed files
	Output string
	// RenameMode determines how files are renamed ("symlink" or "copy")
	RenameMode string
	// Write enables actual file operations; false enables dry-run mode
	Write bool
}

// Entry represents a single file rename operation.
type Entry struct {
	// Destination is the target path for the renamed file
	Destination string
	// Error contains any error encountered while processing this entry
	Error error
	// Source is the original path of the file
	Source string
}

// New creates a new Renamer instance with the given paths, providers, and options.
// It returns an error if no paths or providers are specified.
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

// Run executes the renaming process by scanning paths, generating entries,
// creating directories, and performing rename operations based on the configured mode.
func (r *renamer) Run(o source.Options) (err error) {
	// log output prefix
	prefix := ""
	if !r.o.Write {
		prefix = "[dry-run] "
	}

	// Select the appropriate writer function based on the rename mode
	var w writer
	switch r.o.RenameMode {
	case "symlink":
		w = os.Symlink
	case "copy":
		w = copyFile
	default:
		return fmt.Errorf("unknown rename mode: %s", r.o.RenameMode)
	}

	log.Info().Int("providers", len(r.providers)).Msgf("scanning %d path(s)", len(r.paths))

	// Scan all paths and collect inforations for renaming
	nodes := make(map[string][]source.Node)
	for _, path := range r.paths {
		output := r.o.Output
		if output == "" {
			output = filepath.Dir(filepath.Clean(path))
		}

		n := source.Scan(path, r.providers, o)

		nodes[output] = append(nodes[output], n...)
	}

	if len(nodes) == 0 {
		log.Warn().Msg("no results found")
		return nil
	}

	// Generate rename entries using formatter and collected nodes
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

	// Start renaming by creating necessary uniq directories
	dirCount := 0
	slices.Sort(dirs)
	for _, dir := range slices.Compact(dirs) {
		if r.o.Write {
			err := os.MkdirAll(dir, 0o750)
			if err != nil {
				return fmt.Errorf("failed to create directory %q: %w", dir, err)
			}
		}
		log.Info().Msgf("%s[new directory] -> [%s]", prefix, dir)
		dirCount++
	}
	if dirCount > 0 {
		log.Info().Msgf("%screated %d directories", prefix, dirCount)
	}

	// Rename files
	renamedCount := 0
	uniqEntries := make(map[string]struct{})
	// use index here to be able to set error on the entry
	for index := range entries {
		if entries[index].Error != nil {
			continue
		}

		// Check for duplicate destination paths
		if _, exists := uniqEntries[entries[index].Destination]; exists {
			entries[index].Error = fmt.Errorf("duplicate destination path: %s", entries[index].Destination)
			continue
		}

		// Prepare symlink source if needed
		realSrc := entries[index].Source
		if r.o.RenameMode == "symlink" {
			realSrc, err = getSymlinkSrc(entries[index].Source, entries[index].Destination)
			if err != nil {
				entries[index].Error = err
				continue
			}
		}

		// Perform the write operation
		err := r.write(realSrc, entries[index].Destination, w)
		if err != nil {
			entries[index].Error = err
			continue
		}

		uniqEntries[entries[index].Destination] = struct{}{}

		log.Info().Msgf("%s[%s]\nrenamed to %s[%s]", prefix, entries[index].Source, prefix, entries[index].Destination)
		renamedCount++
	}

	// Print summary of errors and renamed files
	errorsCount := 0
	for _, e := range entries {
		if e.Error == nil {
			continue
		}

		if errors.Is(e.Error, source.ErrExcludedPath) {
			log.Warn().Err(e.Error).Msgf("%s[%s]", prefix, e.Source)
		} else {
			log.Err(e.Error).Msgf("%s[%s]", prefix, e.Source)
			errorsCount++
		}
	}

	e := log.Info()
	if errorsCount > 0 {
		e = log.Warn()
	}
	e.Msgf("%srenamed %d/%d file(s)", prefix, renamedCount, len(entries))

	// TODO: handle non destructive renaming, keeping other files (subtitles, etc)
	// TODO: include directories in the renaming process

	return nil
}

// generateEntry creates an Entry from a source node by formatting the destination path
// based on the node's metadata. It returns the entry and any parent directory to create.
func (r *renamer) generateEntry(node source.Node, output string) (e Entry, dir string) {
	e.Source = node.Path

	if node.Error != nil {
		e.Error = node.Error
		return
	}

	// Call the appropriate formatter method based on the response type
	var components []string
	switch resp := node.Response.(type) {
	case provider.ResponseMovie:
		components = r.o.Formatter.Movie(resp, node)
	case provider.ResponseTV:
		components = r.o.Formatter.TVShow(resp, node)
	case provider.ResponseTVSeason:
		components = r.o.Formatter.TVSeason(resp, node)
	case provider.ResponseTVEpisode:
		components = r.o.Formatter.TVEpisode(resp, node)
	default:
		e.Error = fmt.Errorf("unknown type: %T", node.Response)
		return
	}
	if len(components) == 0 {
		e.Error = fmt.Errorf("no components")
		return
	}

	// Read file extension if not a directory
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

	// Check for duplicate source paths
	// TODO: move this outside of this function, to avoid using r.files state
	if _, exists := r.files[node.Path]; exists {
		e.Error = fmt.Errorf("duplicate source path")
		return
	}

	// Ensure destination path is unique
	// If another file is already using the same destination,
	// attempt deduplication by appending deduplicationSuffix x deduplicationAttemptLimit times
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

// writer is a function type that performs the actual file operation (symlink or copy).
type writer func(string, string) error

// write executes the writer function if Write mode is enabled, otherwise does nothing.
func (r *renamer) write(src, dst string, fn writer) error {
	if !r.o.Write || fn == nil {
		return nil
	}

	if r.o.Force {
		if _, err := os.Lstat(dst); err == nil {
			err = os.Remove(dst)
			if err != nil {
				return fmt.Errorf("failed to remove existing destination %q: %w", dst, err)
			}
		}
	}

	return fn(src, dst)
}

// getSymlinkSrc calculates the relative path from dst to src for creating a symlink.
func getSymlinkSrc(src, dst string) (string, error) {
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path of source %q: %w", src, err)
	}
	absDst, err := filepath.Abs(filepath.Dir(dst))
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path of destination %q: %w", dst, err)
	}
	realSrc, err := filepath.Rel(absDst, absSrc)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path from %q to %q: %w", absDst, absSrc, err)
	}

	return realSrc, nil
}

// copyFile copies a regular file from src to dst.
func copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source file %q: %w", src, err)
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %q: %w", src, err)
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file %q: %w", dst, err)
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return fmt.Errorf("failed to copy from %q to %q: %w", src, dst, err)
}
