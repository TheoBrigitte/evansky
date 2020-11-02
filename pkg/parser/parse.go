package parser

import (
	parse "github.com/middelink/go-parse-torrent-name"
)

func Parse(filename string) (*parse.TorrentInfo, error) {
	return parse.Parse(filename)
}
