module github.com/TheoBrigitte/evansky

go 1.25

require (
	github.com/cyruzin/golang-tmdb v1.8.2
	github.com/docker/go-units v0.5.0
	github.com/middelink/go-parse-torrent-name v0.0.0-20190301154245-3ff4efacd4c4
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.10.1
	github.com/spf13/pflag v1.0.10
	golang.org/x/crypto v0.42.0
)

require (
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/term v0.35.0 // indirect
)

replace github.com/middelink/go-parse-torrent-name => ../go-parse-torrent-name
