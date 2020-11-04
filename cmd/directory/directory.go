package directory

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/TheoBrigitte/evansky/cmd/directory/list"
	"github.com/TheoBrigitte/evansky/cmd/directory/rename"
	"github.com/TheoBrigitte/evansky/cmd/directory/scan"
	"github.com/TheoBrigitte/evansky/cmd/directory/show"
)

// Cmd represents the list command
var Cmd = &cobra.Command{
	Use:               "directory",
	Short:             "manage directories",
	PersistentPreRunE: preventNonSensitiveFS,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := Cmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	Cmd.AddCommand(list.Cmd)
	Cmd.AddCommand(scan.Cmd)
	Cmd.AddCommand(rename.Cmd)
	Cmd.AddCommand(show.Cmd)
}
