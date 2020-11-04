package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/TheoBrigitte/evansky/cmd/cache"
	"github.com/TheoBrigitte/evansky/cmd/completion"
	"github.com/TheoBrigitte/evansky/cmd/directory"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:               "evansky",
	Short:             "media renamer",
	Long:              `Rename media files in order to be detected by media server like Jellyfin.`,
	PersistentPreRunE: logLevel,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(cache.Cmd)
	rootCmd.AddCommand(completion.Cmd)
	rootCmd.AddCommand(directory.Cmd)
}
