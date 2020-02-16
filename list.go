package filter

import (
	"io"
	"os"
	"strings"
)

type List struct {
	Path  string
	Block bool
}

func (l *List) Open() (src io.ReadCloser, err error) {
	if strings.HasPrefix(l.Path, ".") {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		file, err := os.Open(cwd + strings.TrimPrefix(l.Path, "."))
		if err != nil {
			return nil, err
		}
		src = file

	} else {
		file, err := os.Open(l.Path)
		if err != nil {
			return nil, err
		}
		src = file
	}
	return src, nil
}
