package language

import "strings"

// NormalizeLanguage normalizes language string
func NormalizeLanguage(input string) string {
	switch strings.ToLower(input) {
	case "en", "english":
		return "en"
	case "fr", "french", "vf":
		return "fr"
	}

	// Could not normalize language
	return ""
}
