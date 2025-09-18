package tmdb

import (
	"fmt"

	gotmdb "github.com/cyruzin/golang-tmdb"

	"github.com/TheoBrigitte/evansky/pkg/provider"
)

// Client to communicate with tmdb api.
type Client struct {
	client *gotmdb.Client
}

// makeResponse is a generic function to convert an array of api responses, into an array of provider.Responses
func makeResponse[R any, T provider.Response](results []R, f func(R) (T, error)) (provider.Response, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no result")
	}

	return f(results[0])
	//responses := make([]provider.Response, 0, len(results))
	//for _, result := range results {
	//	r, err := f(result)
	//	if err != nil {
	//		return nil, err
	//	}
	//	responses = append(responses, r)
	//}
	//return responses, nil
}
