package language

import (
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
		//WithMinimumRelativeDistance(0.9).
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

func Detect(req provider.Request) (string, float64) {
	if req.Response == nil {
		// Use default language for initial search, to let the provider decide the best match.
		return "en", -1
	}

	lang, confidence := Lingua(req.Query)

	return strings.ToLower(lang), confidence
}
