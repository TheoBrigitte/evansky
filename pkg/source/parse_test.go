package source

import (
	"fmt"
	"regexp"
	"testing"
)

type testCase struct {
	input    string
	expected string
}

var testData = []struct {
	name  string
	regex *regexp.Regexp
	cases []testCase
}{
	{
		name:  "seasonRegex",
		regex: seasonRegex,
		cases: []testCase{
			{input: "Season 1", expected: "1"},
			{input: "S02", expected: "02"},
			{input: "Season 10", expected: "10"},
			{input: "S3", expected: "3"},
			{input: "Season", expected: ""},
			{input: "NoSeasonHere", expected: ""},
			{input: "Season 1 Extra 2", expected: "1"},
		},
	},
	{
		name:  "episodeRegex",
		regex: episodeRegex,
		cases: []testCase{
			{input: "01 - Pilot", expected: "01"},
			{input: "1 - The Beginning", expected: "1"},
			{input: "10 - The Finale", expected: "10"},
			{input: "Episode 2", expected: ""},
			{input: "NoEpisodeHere", expected: ""},
			{input: "03 Another Title", expected: "03"},
		},
	},
}

func TestParse(t *testing.T) {
	for _, data := range testData {
		for i, tc := range data.cases {
			t.Run(fmt.Sprintf("%s_%03d", data.name, i), func(t *testing.T) {
				matches := data.regex.FindStringSubmatch(tc.input)

				var result string
				if len(matches) > 1 {
					result = matches[1]
				}

				if result != tc.expected {
					t.Errorf("For input '%s', expected '%s' but got '%s'", tc.input, tc.expected, result)
				}
			})
		}
	}
}
