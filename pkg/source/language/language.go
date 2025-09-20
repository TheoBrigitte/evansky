// Package language provides language detection capabilities for media files and directories.
// It supports multiple detection methods and maintains language consistency across
// directory tree traversal to improve metadata accuracy.
package language

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/abadojack/whatlanggo"
	"github.com/pemistahl/lingua-go"

	"github.com/TheoBrigitte/evansky/pkg/provider"
)

// languages defines the supported languages for detection.
// Currently supports English, French, German, and Spanish.
var (
	languages = []lingua.Language{
		lingua.English,
		lingua.French,
		lingua.German,
		lingua.Spanish,
	}

	// detector is a pre-configured language detector using the Lingua library.
	// It's configured with a minimum relative distance of 0.9 for higher accuracy.
	detector = lingua.NewLanguageDetectorBuilder().
			FromLanguages(languages...).
			WithMinimumRelativeDistance(0.9).
			Build()
)

// Lingua detects the language of the input text using the Lingua language detection library.
// It returns the ISO 639-1 language code and a confidence score between 0 and 1.
// If no language is detected, it defaults to English.
func Lingua(input string) (string, float64) {
	lang, exists := detector.DetectLanguageOf(input)
	if !exists {
		lang = lingua.English
	}

	confidence := detector.ComputeLanguageConfidence(input, lang)

	//confidenceValues := detector.ComputeLanguageConfidenceValues(input)
	//for _, elem := range confidenceValues {
	//	fmt.Printf("%s: %.2f\n", elem.Language(), elem.Value())
	//}

	return lang.IsoCode639_1().String(), confidence
}

// Whatlanggo detects the language of the input text using the whatlanggo library.
// It returns the ISO 639-1 language code and a confidence score.
// This provides an alternative detection method to Lingua.
func Whatlanggo(input string) (string, float64) {
	lang := whatlanggo.Detect(input)
	return lang.Lang.Iso6391(), float64(lang.Confidence)
}

// Detect determines the appropriate language for a media request based on multiple factors.
// It considers directory contents, parent request context, and defaults appropriately.
// Returns three values:
// - lang: the detected language code for the current request
// - confidence: detection confidence score (or -1 if not applicable)
// - childLang: the detected language for child directories based on their names
func Detect(req provider.Request, entries []os.DirEntry) (string, float64, string) {
	var childLang string
	if len(entries) > 0 {
		// Read all entries, concatenate their names and detect the language from that.
		// This provides language context for child directories.
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			names = append(names, filepath.Base(entry.Name()))
		}

		lang, _ := Lingua(strings.Join(names, "\n"))
		lang = strings.ToLower(lang)

		//slog.Debug("language: detected childs language", "language", lang, "confidence", confidence)
		childLang = lang
		//return strings.ToLower(lang), confidence
	}

	if req.Response == nil {
		// Use default language for initial search, to let the provider decide the best match.
		slog.Debug("language: no parent, using default language", "language", "en")
		return "en", -1, childLang
	}

	// Use previously detected language
	// Having a previous request means we are already down in the tree.
	prevReq := req.Response.GetRequest()
	if prevReq == nil {
		slog.Debug("language: no parent, using default language", "language", "en")
		return "en", -1, childLang
	}

	slog.Debug("language: detected", "language", prevReq.Language)
	return prevReq.Language, -1, childLang

	//lang, confidence := Lingua(req.Query)
	//return strings.ToLower(lang), confidence
}
