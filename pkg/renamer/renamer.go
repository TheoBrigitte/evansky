package renamer

import (
	"fmt"
	"log/slog"

	"github.com/TheoBrigitte/evansky/pkg/provider"
	"github.com/TheoBrigitte/evansky/pkg/renamer/format"
	"github.com/TheoBrigitte/evansky/pkg/source"
)

type Renamer interface {
	Run(dryRun, force bool) error
}

type renamer struct {
	formatter format.Formatter
	paths     []string
	providers []provider.Interface
}

func New(paths []string, providers []provider.Interface, formatter format.Formatter) (Renamer, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("at least one source is required")
	}
	if len(providers) == 0 {
		return nil, fmt.Errorf("at least one provider is required")
	}

	r := &renamer{
		formatter: formatter,
		paths:     paths,
		providers: providers,
	}

	return r, nil
}

func (r *renamer) Run(dryRun, force bool) error {
	o := source.Options{}

	var nodes []source.Node
	for _, path := range r.paths {
		n, err := source.Scan(path, r.providers, o)
		if err != nil {
			slog.Error("scan failed", "path", path, "error", err)
			continue
		}
		nodes = append(nodes, n...)
	}

	if len(nodes) == 0 {
		slog.Warn("no nodes found")
		return nil
	}

	r.printNodes(nodes)

	// TODO: handle non destructive renaming, keeping other files (subtitles, etc)
	// TODO: include directories in the renaming process

	return nil
}

func (r *renamer) printNodes(nodes []source.Node) {
	//w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	//fmt.Fprintln(w, "Path\tName\tYear\tError\tType")

	for _, n := range nodes {
		msg := ""
		if n.Error != nil {
			msg = n.Error.Error()
		} else {

			name := ""
			switch resp := n.Response.(type) {
			case provider.ResponseMovie:
				name = r.formatter.Movie(resp)
			case provider.ResponseTV:
				name = r.formatter.TVShow(resp)
			case provider.ResponseTVSeason:
				name = r.formatter.TVSeason(resp)
			case provider.ResponseTVEpisode:
				name = r.formatter.TVEpisode(resp)
			default:
				name = "unknown type"
			}
			msg = name
		}

		fmt.Printf("[%s] -> [%s]\n", n.Path, msg)

		//fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%T\n", n.Path, name, year, errMsg, n.Response)
	}

	//w.Flush()
}
