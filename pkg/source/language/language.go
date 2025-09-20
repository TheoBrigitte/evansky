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

var (
	languages = []lingua.Language{
		lingua.English,
		lingua.French,
		lingua.German,
		lingua.Spanish,
	}

	detector = lingua.NewLanguageDetectorBuilder().
			FromLanguages(languages...).
			WithMinimumRelativeDistance(0.9).
			Build()
)

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
func Whatlanggo(input string) (string, float64) {
	lang := whatlanggo.Detect(input)
	return lang.Lang.Iso6391(), float64(lang.Confidence)
}

func Detect(req provider.Request, entries []os.DirEntry) (string, float64, string) {
	var childLang string
	if len(entries) > 0 {
		// Read all entries, concatenate their names and detect the language from that.
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
