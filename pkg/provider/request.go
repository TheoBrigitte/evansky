package provider

import (
	"fmt"
	"io/fs"

	"github.com/TheoBrigitte/evansky/pkg/parser"
)

// Request represents a search request to a provider.
type Request struct {
	Query               string
	Year                int
	QueryLanguage       string
	DestinationLanguage string
	Info                parser.Info
	Entry               fs.DirEntry

	Response Response
}

func (r Request) String() string {
	return fmt.Sprintf("Request{Query: %q, Year: %d, Language: %q, Reponse: %+v}", r.Query, r.Year, r.QueryLanguage, r.Response)
}
