package rename

type Flags struct {
	excludeGlob        []string
	excludeRegex       string
	includeGlob        []string
	includeRegex       string
	force              bool
	language           string
	mediaExtensions    []string
	output             string
	query              string
	queryLanguage      string
	renameMode         string
	stripComponents    int
	subtitleExtensions []string
	titleRegex         string
	skipExisting       bool
	write              bool
}

func NewFlags() *Flags {
	return &Flags{}
}
