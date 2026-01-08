package parsetorrentname

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
)

var patterns = []struct {
	name string
	// Use the last matching pattern. E.g. Year.
	last bool
	kind reflect.Kind
	// REs need to have 2 sub expressions (groups), the first one is "raw", and
	// the second one for the "clean" value.
	// E.g. Epiode matching on "S01E18" will result in: raw = "E18", clean = "18".
	re *regexp.Regexp
}{
	// Season in S01E01 format (case insensitive)
	{"season", false, reflect.Int, regexp.MustCompile(`(?i)(s([0-9]{1,}))e?`)},
	// Season in 1x01 format (case insensitive)
	{"season", false, reflect.Int, regexp.MustCompile(`(?i)(([0-9]{1,}))x`)},
	// Episode in 1x01 format (case insensitive)
	{"episode", false, reflect.Int, regexp.MustCompile(`(?i)([x]([0-9]{2})(?:[^\w]|$))`)},
	// Episode in S01E01 format (case insensitive)
	{"episode", false, reflect.Int, regexp.MustCompile(`(?i)([e]([0-9]{2,})(?:[^\w]|$))`)},
	// Episode in - 01 format (case insensitive)
	{"episode", false, reflect.Int, regexp.MustCompile(`(-\s+([0-9]{2,})(?:[^\w]|$))`)},
	// Year ranges and take the first year, e.g. 1989-2016 => 1989
	{"year", true, reflect.Int, regexp.MustCompile(`\b(((?:19[0-9]|20[0-9])[0-9])-(?:19[0-9]|20[0-9])[0-9])\b`)},
	// Years from 1900 to 2099
	{"year", true, reflect.Int, regexp.MustCompile(`\b(((?:19[0-9]|20[0-9])[0-9]))\b`)},
	// Resolution like 720p, 1080p, 2160p
	{"resolution", false, reflect.String, regexp.MustCompile(`\b(([0-9]{3,4}p))\b`)},
	// Quality like HDTS, DVDRip, BluRay, WEB-DL, CAM, HDRip, etc.
	{"quality", false, reflect.String, regexp.MustCompile(`(?i)\b(((?:PPV\.)?[HP]DTV|(?:HD)?CAM|B[DR]Rip|(?:HD-?)?TS|(?:PPV )?WEB-?DL(?: DVDRip)?|HDRip|DVDRip|DVDRIP|CamRip|W[EB]BRip|BluRay|DvDScr|telesync|WEB))\b`)},
	// Codec like x264, x265, h264, h265, XviD
	{"codec", false, reflect.String, regexp.MustCompile(`(?i)\b((xvid|[hx]\.?26[45]))\b`)},
	// Audio like MP3, DD5.1, Dual-Audio, LiNE, DTS, AAC-LC, AC3.5.1
	{"audio", false, reflect.String, regexp.MustCompile(`(?i)\b((MP3|DD5\.?1|Dual[\- ]Audio|LiNE|DTS|AAC[.-]LC|AAC(?:\.?2\.0)?|AC3(?:\W{0,3}5\.1)?))\b`)},
	// Region like R1, R2, R3, R4, R5, R6
	{"region", false, reflect.String, regexp.MustCompile(`(?i)\b(R([0-9]))\b`)},
	// Size like 700MB, 1.4GB, 2GB
	{"size", false, reflect.String, regexp.MustCompile(`(?i)\b((\d+(?:\.\d+)?(?:GB|MB)))\b`)},
	// Website or release source at start of the string and enclosed in square brackets, e.g. [ www.Speed.cd ], [HorribleSubs]
	{"website", false, reflect.String, regexp.MustCompile(`^(\[ ?([^\]]+?) ?\])`)},
	// Language like VO, VOSTFR, MULTI
	{"language", false, reflect.String, regexp.MustCompile(`(?i)\b((VO|VOSTFR|VF|VFF|VF2|MULTI))\b`)},
	// Language like rus.eng, ita.eng
	{"language", false, reflect.String, regexp.MustCompile(`(?i)\b((en(?:glish)?|fr(?:ench)?|rus\.eng|ita\.eng))\b`)},
	{"sbs", false, reflect.String, regexp.MustCompile(`(?i)\b(((?:Half-)?SBS))\b`)},
	// Container like mkv, avi, mp4
	{"container", false, reflect.String, regexp.MustCompile(`(?i)\b((mkv|avi|mp4))\b`)},

	// Group like YIFY, RARBG, SPARKS, FoV, KILLERS
	{"group", false, reflect.String, regexp.MustCompile(`(- ?(.+?))+(?:\.\w+)?$`)},

	{"extended", false, reflect.Bool, regexp.MustCompile(`(?i)\b(EXTENDED(:?.CUT)?)\b`)},
	{"hardcoded", false, reflect.Bool, regexp.MustCompile(`(?i)\b((HC))\b`)},
	{"proper", false, reflect.Bool, regexp.MustCompile(`(?i)\b((PROPER))\b`)},
	{"repack", false, reflect.Bool, regexp.MustCompile(`(?i)\b((REPACK))\b`)},
	{"widescreen", false, reflect.Bool, regexp.MustCompile(`(?i)\b((WS))\b`)},
	{"unrated", false, reflect.Bool, regexp.MustCompile(`(?i)\b((UNRATED))\b`)},
	{"threeD", false, reflect.Bool, regexp.MustCompile(`(?i)\b((3D))\b`)},
}

func init() {
	for _, pat := range patterns {
		if pat.re.NumSubexp() != 2 {
			fmt.Printf("Pattern %q does not have enough capture groups. want 2, got %d\n", pat.name, pat.re.NumSubexp())
			os.Exit(1)
		}
	}
}
