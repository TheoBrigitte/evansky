// Package rename implements the "rename" command, which renames and organizes directory content based on metadata providers and user-defined options.
package rename

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/TheoBrigitte/evansky/pkg/provider/register"
	"github.com/TheoBrigitte/evansky/pkg/renamer"
	"github.com/TheoBrigitte/evansky/pkg/renamer/format"
	"github.com/TheoBrigitte/evansky/pkg/source"

	"github.com/spf13/cobra"
)

var (
	Cmd = &cobra.Command{
		Use:   "rename [flags] <file | directory>...",
		Short: "rename directory content",
		Long: `Rename and organize directory content` +
			`--exclude and --exclude-regex are anded.`,
		RunE: runner,
		Args: cobra.MinimumNArgs(1),
	}

	flags *Flags
)

func init() {
	flags = NewFlags()

	Cmd.PersistentFlags().StringVar(&flags.excludeGlob, "exclude", "", "exclude files matching the given glob pattern")
	Cmd.PersistentFlags().StringVar(&flags.excludeRegex, "exclude-regex", "", "exclude files matching the given regular expression")
	Cmd.PersistentFlags().StringVar(&flags.includeRegex, "include-regex", "", "include files matching the given regular expression")
	Cmd.PersistentFlags().BoolVarP(&flags.force, "force", "f", false, "overwrite existing destination files")
	Cmd.PersistentFlags().StringVar(&flags.language, "language", "en", "language used for destination names (ISO 639-1 code)")
	Cmd.PersistentFlags().StringSliceVar(&flags.mediaExtensions, "media-ext", []string{"mkv", "mp4", "avi", "mov", "wmv", "flv", "mpg", "mpeg"}, "media file extensions to consider")
	Cmd.PersistentFlags().StringVarP(&flags.output, "output", "o", "", "output directory (default: same as source)")
	Cmd.PersistentFlags().StringVar(&flags.query, "query", "", "search query override")
	Cmd.PersistentFlags().StringVar(&flags.queryLanguage, "query-language", "", "language query override")
	Cmd.PersistentFlags().StringVar(&flags.renameMode, "mode", "symlink", "rename mode: symlink, hardlink, copy, move")
	Cmd.PersistentFlags().IntVar(&flags.stripComponents, "strip-components", 0, "number of leading path components to strip from source paths")
	Cmd.PersistentFlags().StringVar(&flags.titleRegex, "title-regex", "", "regular expression to extract title from file or directory name")
	Cmd.PersistentFlags().StringSliceVar(&flags.subtitleExtensions, "subtitle-ext", []string{"srt", "idx", "sub"}, "subtitles extensions to consider")
	Cmd.PersistentFlags().BoolVar(&flags.skipExisting, "skip-existing", false, "skip renaming if destination dir already exists")
	Cmd.PersistentFlags().BoolVar(&flags.write, "write", false, "actually perform the rename operation (default: false)")

	register.Initialize(Cmd)
}

func runner(cmd *cobra.Command, args []string) error {
	providers, err := register.GetProviders()
	if err != nil {
		return err
	}

	formatter := format.NewJellyfinFormatter()

	if flags.output != "" {
		info, err := os.Lstat(flags.output)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if err == nil && !info.IsDir() {
			return fmt.Errorf("output is not a directory: %s", flags.output)
		}
		flags.output = filepath.Clean(flags.output)
	}

	renameOptions := renamer.Options{
		Force:        flags.force,
		Formatter:    formatter,
		Output:       flags.output,
		RenameMode:   flags.renameMode,
		SkipExisting: flags.skipExisting,
	}
	if flags.write {
		renameOptions.Write = true
	}

	r, err := renamer.New(args, providers, renameOptions)
	if err != nil {
		return err
	}

	sourceOptions := source.Options{
		Query:           flags.query,
		QueryLanguage:   flags.queryLanguage,
		Language:        flags.language,
		ExcludeGlob:     flags.excludeGlob,
		ExcludeRegex:    flags.excludeRegex,
		IncludeRegex:    flags.includeRegex,
		MediaExts:       flags.mediaExtensions,
		SubtitleExts:    flags.subtitleExtensions,
		StripComponents: flags.stripComponents,
		TitleRegex:      flags.titleRegex,
	}

	return r.Run(sourceOptions)
}
