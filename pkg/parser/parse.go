// Package parser provides functions to parse torrent file names and extract relevant information such as title, year, season, episode, etc. It uses the go-parse-torrent-name library for parsing.
package parser

import (
	parse "github.com/middelink/go-parse-torrent-name"
)

type Info = parse.TorrentInfo

func Parse(filename string) (*Info, error) {
	return parse.Parse(filename)
}

func CleanTitle(title string) string {
	return parse.CleanTitle(title)
}
