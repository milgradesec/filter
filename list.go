package filter

import (
	"io"
	"os"
)

type list struct {
	Path  string
	Block bool
}

func (l *list) Read() (src io.ReadCloser, err error) {
	f, err := os.Open(l.Path)
	if err != nil {
		return nil, err
	}
	return f, nil
}
