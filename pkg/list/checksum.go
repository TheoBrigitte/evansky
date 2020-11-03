package list

import (
	"crypto/sha256"
	"os"
)

func Sum(list []os.FileInfo) ([]byte, error) {
	h := sha256.New()

	for i, f := range list {
		_, err := h.Write([]byte(f.Name()))
		if err != nil {
			return nil, err
		}

		if i < len(list)-1 {
			_, err := h.Write([]byte(","))
			if err != nil {
				return nil, err
			}
		}
	}

	return h.Sum(nil), nil
}
