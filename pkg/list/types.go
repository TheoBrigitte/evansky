package list

import (
	"os"
)

type Result struct {
	Files         int    `json:"files"`
	FilesChecksum string `json:"filesChecksum"`
	Path          string `json:"path"`
	PathChecksum  string `json:"pathChecksum"`
}

type Lister struct {
	files         []os.FileInfo
	filesChecksum string
	path          string
	pathChecksum  string
}

func (l Lister) Path() string {
	return l.path
}

func (l Lister) PathChecksum() string {
	return l.pathChecksum
}

func (l Lister) Files() []os.FileInfo {
	return l.files
}

func (l Lister) FilesChecksum() string {
	return l.filesChecksum
}
