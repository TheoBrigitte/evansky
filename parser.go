package parsetorrentname

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// TorrentInfo is the resulting structure returned by Parse
type TorrentInfo struct {
	Title      string
	Season     int    `json:"season,omitempty"`
	Episode    int    `json:"episode,omitempty"`
	Year       int    `json:"year,omitempty"`
	Resolution string `json:"resolution,omitempty"`
	Quality    string `json:"quality,omitempty"`
	Codec      string `json:"codec,omitempty"`
	Audio      string `json:"audio,omitempty"`
	Group      string `json:"group,omitempty"`
	Region     string `json:"region,omitempty"`
	Extended   bool   `json:"extended,omitempty"`
	Hardcoded  bool   `json:"hardcoded,omitempty"`
	Proper     bool   `json:"proper,omitempty"`
	Repack     bool   `json:"repack,omitempty"`
	Container  string `json:"container,omitempty"`
	Widescreen bool   `json:"widescreen,omitempty"`
	Website    string `json:"website,omitempty"`
	Language   string `json:"language,omitempty"`
	Sbs        string `json:"sbs,omitempty"`
	Unrated    bool   `json:"unrated,omitempty"`
	Size       string `json:"size,omitempty"`
	ThreeD     bool   `json:"3d,omitempty"`
}

var (
	titleCaser   = cases.Title(language.English, cases.NoLower)

	debug = true
)

func setField(tor *TorrentInfo, field, raw, val string) {
	ttor := reflect.TypeOf(tor)
	torV := reflect.ValueOf(tor)
	field = titleCaser.String(field)
	v, _ := ttor.Elem().FieldByName(field)
	if debug {
		fmt.Printf("    field=%v, type=%+v, value=%v\n", field, v.Type, val)
	}
	switch v.Type.Kind() {
	case reflect.Bool:
		torV.Elem().FieldByName(field).SetBool(true)
	case reflect.Int:
		clean, _ := strconv.ParseInt(val, 10, 64)
		torV.Elem().FieldByName(field).SetInt(clean)
	case reflect.Uint:
		clean, _ := strconv.ParseUint(val, 10, 64)
		torV.Elem().FieldByName(field).SetUint(clean)
	case reflect.String:
		torV.Elem().FieldByName(field).SetString(val)
	}
}

// Parse breaks up the given filename in TorrentInfo
func Parse(filename string) (*TorrentInfo, error) {
	// tor holds the resulting parsed information
	tor := &TorrentInfo{}

	cleanName := strings.Replace(filename, "_", " ", -1)
	// titleStartIndex and titleEndIndex hold the indexes for the title extraction
	titleStartIndex, titleEndIndex := 0, len(filename)
	if debug {
		fmt.Printf("filename %q\n", filename)
	}

	// Process all patterns
	patternMatches := make(map[string]string)
	for _, pattern := range patterns {
		// Skip if a similar pattern already matched
		if _, ok := patternMatches[pattern.name]; ok {
			continue
		}
		//// If the value already exists, it is no longer updated. see golden_file_083.json
		//if pattern.name == "episode" && tor.Episode != 0 {
		//	continue
		//}

		// Match all occurences of the pattern against the cleanName
		groups := pattern.re.FindAllStringSubmatch(cleanName, -1)
		if debug {
			fmt.Printf("  %s: pattern:%q match:%#v\n", pattern.name, pattern.re, groups)
		}

		if len(groups) <= 0 {
			// No match for this pattern
			continue
		}

		// Select match group
		matches := groups[0]
		if pattern.last {
			// Take last match group.
			matches = groups[len(groups)-1]
		}

		if len(matches) <= 2 {
			// No matches in this group
			continue
		}

		// Skip matches overlap, which occurs when the current match contains part of an already matched pattern
		if containsPartOf(matches[2], patternMatches) {
			continue
		}

		// Update title index
		index := strings.Index(cleanName, matches[1])
		if index == 0 {
			// Move title start index after this match
			titleStartIndex = len(matches[1])
			if debug {
				fmt.Printf("    startIndex moved to %d [%q]\n", titleStartIndex, filename[titleStartIndex:titleEndIndex])
			}
		} else if index < titleEndIndex {
			// Move title end index before this match
			titleEndIndex = index
			if debug {
				fmt.Printf("    endIndex moved to %d [%q]\n", titleEndIndex, filename[titleStartIndex:titleEndIndex])
			}
		}

		setField(tor, pattern.name, matches[1], matches[2])

		// Set pattern as already matched
		patternMatches[pattern.name] = matches[2]
	}

	// Start process for title
	// fmt.Println("  title: <internal>")
	if titleStartIndex > titleEndIndex {
		titleStartIndex = 0
	}

	// Take the first filename part before a '(' as title
	parts := strings.Split(filename[titleStartIndex:titleEndIndex], "(")
	if len(parts) < 1 {
		return tor, nil
	}
	raw := parts[0]

	// Set title
	setField(tor, "title", raw, CleanTitle(raw))

	return tor, nil
}

func CleanTitle(raw string) string {
	// Remove leading and trailing spaces and dashes
	cleanName := strings.Trim(raw, " -")

	if strings.ContainsRune(cleanName, '.') && !strings.ContainsRune(cleanName, ' ') {
		// If there are dots but no spaces, replace dots with spaces
		// examples:
		//          rename "Doctor.Who." to "Doctor Who"
		//   do not rename "Marvels Agents of S.H.I.E.L.D."
		cleanName = strings.ReplaceAll(cleanName, ".", " ")
	}

	// Replace underscores with spaces
	cleanName = strings.ReplaceAll(cleanName, "_", " ")

	// cleanName = re.sub('([\[\(_]|- )$', '', cleanName).strip()

	return strings.TrimSpace(cleanName)
}

// containsPartOf returns true if s contains part of any of the values in patternMatches
// A part of a value is contained when its first or last 3 characters are found in s.
// Patterns for year, season, and episode are ignored, this is because they are highly likely to return a false positive.
func containsPartOf(s string, patternMatches map[string]string) bool {
	for name, value := range patternMatches {
		switch name {
		case "year", "season", "episode":
			continue
		}

		if len(value) > 3 {
			if strings.Contains(s, value[:3]) {
				return true
			}
			if strings.Contains(s, value[len(value)-4:]) {
				return true
			}
		}
		if strings.Contains(s, value) {
			return true
		}
	}

	return false
}
