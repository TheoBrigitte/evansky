package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func persistentPreRunE(cmd *cobra.Command, args []string) error {
	err := logLevel(cmd, args)
	if err != nil {
		return err
	}

	err = preventNonSensitiveFS(cmd, args)
	if err != nil {
		return err
	}

	return nil
}

// logLevel set the level of the logger.
func logLevel(cmd *cobra.Command, args []string) error {
	level, err := log.ParseLevel(cmd.Flag("log-level").Value.String())
	if err != nil {
		return err
	}
	log.SetLevel(level)

	return nil
}

// preventNonSensitiveFS check that we are running on a case sensitive filesystem.
// This is needed otherwise we might run into cases where we cannot move folder because of case sensitivity.
// e.g. $ mv test /tmp/
//      this does not work when the layout contains:
//      $ ls /tmp
//      Test
func preventNonSensitiveFS(cmd *cobra.Command, args []string) error {
	p, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	ok, err := isCaseSensitiveFilesystem(p)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("non sensitive filesystem: aborting")
	}

	return nil
}

// IsCaseSensitiveFilesystem determines if the filesystem where dir
// exists is case sensitive or not.
//
// CAVEAT: this function works by taking the last component of the given
// path and flipping the case of the first letter for which case
// flipping is a reversible operation (/foo/Bar → /foo/bar), then
// testing for the existence of the new filename. There are two
// possibilities:
//
// 1. The alternate filename does not exist. We can conclude that the
// filesystem is case sensitive.
//
// 2. The filename happens to exist. We have to test if the two files
// are the same file (case insensitive file system) or different ones
// (case sensitive filesystem).
//
// If the input directory is such that the last component is composed
// exclusively of case-less codepoints (e.g.  numbers), this function will
// return false.
//
// taken from: https://github.com/golang/dep/blob/f13583b555deaa6742f141a9c1185af947720d60/internal/fs/fs.go#L215
func isCaseSensitiveFilesystem(dir string) (bool, error) {
	alt := filepath.Join(filepath.Dir(dir), genTestFilename(filepath.Base(dir)))

	dInfo, err := os.Stat(dir)
	if err != nil {
		return false, errors.Wrap(err, "could not determine the case-sensitivity of the filesystem")
	}

	aInfo, err := os.Stat(alt)
	if err != nil {
		// If the file doesn't exists, assume we are on a case-sensitive filesystem.
		if os.IsNotExist(err) {
			return true, nil
		}

		return false, errors.Wrap(err, "could not determine the case-sensitivity of the filesystem")
	}

	return !os.SameFile(dInfo, aInfo), nil
}

// genTestFilename returns a string with at most one rune case-flipped.
//
// The transformation is applied only to the first rune that can be
// reversibly case-flipped, meaning:
//
// * A lowercase rune for which it's true that lower(upper(r)) == r
// * An uppercase rune for which it's true that upper(lower(r)) == r
//
// All the other runes are left intact.
func genTestFilename(str string) string {
	flip := true
	return strings.Map(func(r rune) rune {
		if flip {
			if unicode.IsLower(r) {
				u := unicode.ToUpper(r)
				if unicode.ToLower(u) == r {
					r = u
					flip = false
				}
			} else if unicode.IsUpper(r) {
				l := unicode.ToLower(r)
				if unicode.ToUpper(l) == r {
					r = l
					flip = false
				}
			}
		}
		return r
	}, str)
}
