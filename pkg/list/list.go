package list

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func New(path string) (*Lister, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	pathChecksum := fmt.Sprintf("%x", sha256.Sum256([]byte(absolutePath)))

	l := &Lister{
		path:         absolutePath,
		pathChecksum: pathChecksum,
	}

	log.Debugf("lister for %s (sha256: %s)\n", l.path, l.pathChecksum)

	return l, nil
}

func (l *Lister) List() (*Result, error) {
	files, err := ioutil.ReadDir(l.Path())
	if err != nil {
		return nil, err
	}
	s, err := Sum(files)
	if err != nil {
		return nil, err
	}

	filesChecksum := fmt.Sprintf("%x", s)

	l.files = files
	l.filesChecksum = filesChecksum

	lr := &Result{
		Files:         len(l.files),
		FilesChecksum: l.filesChecksum,
		Path:          l.path,
		PathChecksum:  l.pathChecksum,
	}

	log.Debugf("listed %d file(s) (sha256: %s)\n", lr.Files, lr.FilesChecksum)

	return lr, nil
}
