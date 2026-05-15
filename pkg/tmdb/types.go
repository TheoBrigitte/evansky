package tmdb

import (
	"context"

	"github.com/golusoris/goenvoy/metadata/video/tmdb"
)

// Client to communicate with tmdb api.
type Client struct {
	client *tmdb.Client
	ctx    context.Context
}
