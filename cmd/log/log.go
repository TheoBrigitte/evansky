// Package log provides common logging configuration for cobra commands.
package log

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	// Flag variables
	logLevel      string
	logWithSource bool

	// logLevels maps string representations of log levels to zerolog.Level values.
	logLevels = map[string]zerolog.Level{
		"debug": zerolog.DebugLevel,
		"info":  zerolog.InfoLevel,
		"warn":  zerolog.WarnLevel,
		"error": zerolog.ErrorLevel,
	}
)

// AddFlags adds logging flags to the provided cobra command and sets up a PersistentPreRunE to configure logging.
func AddFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", fmt.Sprintf("set log level (%s)", allLogLevels()))
	cmd.PersistentFlags().BoolVar(&logWithSource, "log-with-source", false, "include source file and line number in log messages")
	cmd.PersistentPreRunE = setLogLevel
}

// allLogLevels returns a comma-separated string of all valid log levels.
func allLogLevels() string {
	levels := make([]string, 0, len(logLevels))
	for level := range logLevels {
		levels = append(levels, level)
	}
	return strings.Join(levels, ", ")
}

// setLogLevel set the level of the zerolog logger based on the logLevel variable.
// It also configures the logger to output to stderr in a human-friendly format.
func setLogLevel(cmd *cobra.Command, args []string) error {
	level, err := parseLogLevel(logLevel)
	if err != nil {
		return err
	}

	zerolog.SetGlobalLevel(level)

	// Set log to be printed on stderr with human-friendly format
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:          os.Stderr,
		PartsExclude: []string{"time"},
		FieldsOrder:  []string{"level", "message", "destination", "source"},
	})

	return nil
}

// parseLogLevel converts a string log level to a zerolog.Level. It returns an error if the log level is invalid.
func parseLogLevel(logLevelStr string) (zerolog.Level, error) {
	level, ok := logLevels[logLevelStr]
	if !ok {
		return -1, fmt.Errorf("invalid log level: %s", logLevelStr)
	}
	return level, nil
}
