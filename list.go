package filter

import (
	"net/http"
	"os"
	"strings"
)

func (f *Filter) Load() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	for list := range f.lists {
		log.Debugf("Loading list %v", list)

		if strings.HasPrefix(list, "http") {
			resp, err := http.Get(list)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			_ = resp

		} else if strings.HasPrefix(list, ".") {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			file, err := os.Open(cwd + list)
			if err != nil {
				return err
			}
			defer file.Close()

		} else {
			file, err := os.Open(list)
			if err != nil {
				return err
			}
			defer file.Close()
		}
	}
	return nil
}
