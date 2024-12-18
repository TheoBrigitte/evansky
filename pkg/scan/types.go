package scan

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/TheoBrigitte/evansky/pkg/movie"
	"github.com/TheoBrigitte/evansky/pkg/tmdb"
)

type Scanner struct {
	client *tmdb.Client
	noAPI  bool
}

type Config struct {
	APIKey string
	NoAPI  bool
}

type Results struct {
	Total    int                    `json:"total"`
	Found    int                    `json:"found"`
	Failures int                    `json:"failures"`
	Results  map[string]movie.Multi `json:"results"`
	Failed   map[string]movie.Multi `json:"failed"`
}

func NewResults() *Results {
	r := &Results{}
	r.Results = make(map[string]movie.Multi)
	r.Failed = make(map[string]movie.Multi)

	return r
}

func (r *Results) IsComplete() bool {
	return r.Found == r.Total
}

func (r *Results) CompletePercentage() string {
	return fmt.Sprintf("%d%%", r.Found*100.0/r.Total)
}

func (r *Results) Print(output io.Writer, verbose bool) {
	if len(r.Results) > 0 || len(r.Failed) > 0 {
		w := tabwriter.NewWriter(output, 0, 0, 7, ' ', tabwriter.AlignRight)

		fmt.Fprintf(w, "original\tnew\t")
		if verbose {
			fmt.Fprintf(w, "id\tisDir\t")
		}
		fmt.Fprintln(w, "")

		fmt.Fprintf(w, "--------\t---\t")
		if verbose {
			fmt.Fprintf(w, "--\t-----\t")
		}
		fmt.Fprintln(w, "")

		for name, r := range r.Results {
			fmt.Fprintf(w, "%s\t%s\t", name, r.Path)
			if verbose {
				fmt.Fprintf(w, "%d\t%t\t", r.ID, r.IsDir)
			}

			fmt.Fprintln(w, "")
		}

		fmt.Fprintln(w, "\t\t")
		fmt.Fprintf(w, "failed\terror\t")
		if verbose {
			fmt.Fprintf(w, "title\tyear\t")
		}
		fmt.Fprintln(w, "")

		fmt.Fprintf(w, "------\t-----\t")
		if verbose {
			fmt.Fprintf(w, "-----\t----\t")
		}
		fmt.Fprintln(w, "")
		for name, r := range r.Failed {
			fmt.Fprintf(w, "%s\t%s\t", name, r.Error)
			if verbose {
				fmt.Fprintf(w, "%s\t%d\t", r.Title, r.Year)
			}
			fmt.Fprintln(w, "")
		}
		w.Flush()
	} else {
		fmt.Fprintln(output, "no results")
	}

	fmt.Fprintf(output, "%d/%d result(s)  %d failure(s)  %s complete\n", r.Found, r.Total, r.Failures, r.CompletePercentage())
}
