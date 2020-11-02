package list

import (
	"crypto/md5"
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

	pathChecksum := fmt.Sprintf("%x", md5.Sum([]byte(absolutePath)))

	l := &Lister{
		path:         absolutePath,
		pathChecksum: pathChecksum,
	}

	log.Debugf("lister for %s (md5: %s)\n", l.path, l.pathChecksum)

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

	log.Debugf("listed %d file(s) (md5: %s)\n", lr.Files, lr.FilesChecksum)

	return lr, nil
}
