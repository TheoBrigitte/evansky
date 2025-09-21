package source

import (
	"fmt"
	"regexp"
	"strconv"
)

var (
	// seasonRegex is a regular expression used to extract numeric values
	// from season directory names for season number detection.
	// It works for common season directory naming patterns like "Season 1", "S02", etc.
	seasonRegex = regexp.MustCompile(`([0-9]+)`)
	// episodeRegex is a regular expression used to extract the episode number
	// from episode file names for episode number detection.
	// It works for common episode naming patterns like "01 - Episode Title", "1 - Episode Title", etc.
	episodeRegex = regexp.MustCompile(`(^[0-9]{1,})\s`)
)

// extractNumber extracts a numeric value from the input string using the provided regex.
// It extracts all numeric sequences from the name and returns the first occurrence as an integer.
func extractNumber(input string, regex *regexp.Regexp) (int, error) {
	matches := regex.FindStringSubmatch(input)
	if len(matches) > 1 {
		// Convert the first match to an integer.
		number, err := strconv.Atoi(matches[1])
		if err != nil {
			return -1, err
		}
		return number, nil
	}

	return -1, fmt.Errorf("no match found")
}
