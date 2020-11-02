package cache

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/TheoBrigitte/evansky/cmd/cache/clean"
	"github.com/TheoBrigitte/evansky/cmd/cache/list"
)

// Cmd represents the list command
var Cmd = &cobra.Command{
	Use:   "cache",
	Short: "manage cache",
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
	Cmd.AddCommand(clean.Cmd)
}
