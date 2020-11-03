package cmd

import (
	log "github.com/sirupsen/logrus"
)

func init() {
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "show less informations")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "show more informations")
	rootCmd.PersistentFlags().StringP("log-level", "l", log.InfoLevel.String(), "log level")
}
