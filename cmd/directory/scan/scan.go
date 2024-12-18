package scan

import (
	"fmt"
	"os"

	"github.com/TheoBrigitte/evansky/pkg/cache"
	"github.com/TheoBrigitte/evansky/pkg/list"
	"github.com/TheoBrigitte/evansky/pkg/scan"

	"github.com/spf13/cobra"
)

// Cmd represents the scan command
var (
	Cmd = &cobra.Command{
		Use:   "scan",
		Short: "scan directory contents",
		Long:  `List content of directory, cache checksum and modification to be made`,
		RunE:  runner,
		Args:  cobra.ExactValidArgs(1),
	}

	apiKey      string
	appendMode  bool
	interactive bool
	noAPI       bool
	force       bool
)

func init() {
	Cmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "tmdb api key")
	Cmd.PersistentFlags().BoolVarP(&appendMode, "append", "a", false, "append mode")
	Cmd.PersistentFlags().BoolVarP(&interactive, "interactive", "i", false, "interactive mode")
	Cmd.PersistentFlags().BoolVarP(&noAPI, "no-api", "n", false, "no api call (not recommended)")
	Cmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "skip confirmation")
}

func runner(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("directory missing")
	}

	// Initializations
	fmt.Printf("scanning %s\n", args[0])
	cfg := scan.Config{
		APIKey: apiKey,
		NoAPI:  noAPI,
	}
	scanner, err := scan.New(cfg)
	if err != nil {
		return err
	}

	lister, err := list.New(args[0])
	if err != nil {
		return err
	}

	c, err := cache.New(lister.PathChecksum())
	if err != nil {
		return err
	}

	// List files
	result, err := lister.List()
	if err != nil {
		return err
	}

	// Check cache status
	status, err := c.Status(lister.FilesChecksum())
	if err != nil {
		return err
	}

	var current *scan.Results
	s, err := c.GetScan()
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}
	if appendMode {
		current = s
		appendMode = !current.IsComplete()
	}

	var verb = "cached"
	if force || appendMode || s == nil || status == cache.DoNotExists || status == cache.Changed {
		err = c.StoreList(result)
		if err != nil {
			return err
		}

		// Scan files
		results, err := scanner.Scan(lister.Files(), interactive, current)
		if err != nil {
			return err
		}

		err = c.StoreScan(results)
		if err != nil {
			return err
		}
		verb = "scanned"
	}

	results, err := c.GetScan()
	if err != nil {
		return err
	}
	fmt.Printf("%s %d file(s), found %d result(s), %d failure(s)\n", verb, results.Total, results.Found, results.Failures)

	return nil
}
