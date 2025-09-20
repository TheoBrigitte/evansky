package provider

import (
	"fmt"

	"github.com/TheoBrigitte/evansky/pkg/parser"
)

// Request represents a search request to a provider.
type Request struct {
	Query    string
	Year     int
	Language string
	Info     parser.Info

	Response Response
}

func (r Request) String() string {
	return fmt.Sprintf("Request{Query: %q, Year: %d, Language: %q, Reponse: %+v}", r.Query, r.Year, r.Language, r.Response)
}
