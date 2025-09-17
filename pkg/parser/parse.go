package parser

import (
	parse "github.com/middelink/go-parse-torrent-name"
)

type Info = parse.TorrentInfo

func Parse(filename string) (*Info, error) {
	return parse.Parse(filename)
}
