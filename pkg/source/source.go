package source

import "os"

type Interface interface {
}

type source struct {
	info os.FileInfo
}

func New(path string) (Interface, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	s := &source{info: info}

	return s, err
}
