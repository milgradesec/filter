package filter

import (
	"bufio"
	"errors"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type list struct {
	hashtable  map[string]struct{}
	prefixes   []string
	suffixes   []string
	subStrings []string
	regexes    []*regexp.Regexp
}

func NewList() *list {
	return &list{
		hashtable: make(map[string]struct{}),
	}
}

var regexpOps = []string{"[", "]", "(", ")", "|", "?",
	"+", "$", "{", "}", "^"}

func (l *list) Read(input io.ReadCloser) error {
	if input == nil {
		return errors.New("invalid list source")
	}
	defer input.Close()

	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		for _, op := range regexpOps {
			if strings.Contains(line, op) {
				r, err := regexp.Compile(line)
				if err != nil {
					return err
				}
				l.regexes = append(l.regexes, r)
				break
			}
		}
		if strings.Contains(line, "*") {
			if strings.HasSuffix(line, "*") && strings.HasPrefix(line, "*") {
				qname := strings.TrimPrefix(line, "*")
				qname = strings.TrimSuffix(qname, "*")
				l.subStrings = append(l.subStrings, qname)
			}
			if strings.HasSuffix(scanner.Text(), "*") {
				domain := strings.TrimSuffix(line, "*")
				l.prefixes = append(l.prefixes, domain)
			}
			if strings.HasPrefix(scanner.Text(), "*") {
				domain := strings.TrimPrefix(line, "*")
				l.suffixes = append(l.suffixes, domain)
			}
		} else {
			l.hashtable[line] = struct{}{}
		}

		if scanner.Err() != nil {
			return scanner.Err()
		}
	}
	return nil
}

func (l *list) Match(str string) bool {
	_, q := l.hashtable[str]
	if q {
		return true
	}
	for _, prefix := range l.prefixes {
		if strings.HasPrefix(str, prefix) {
			return true
		}
	}
	for _, suffix := range l.suffixes {
		if strings.HasSuffix(str, suffix) {
			return true
		}
		if str == strings.TrimPrefix(suffix, ".") {
			return true
		}
	}
	for _, substr := range l.subStrings {
		if strings.Contains(str, substr) {
			return true
		}
	}
	for _, regex := range l.regexes {
		if regex.MatchString(str) {
			return true
		}
	}
	return false
}

func (f *Filter) Load() error {
	whitelist := NewList()
	blocklist := NewList()

	for list, block := range f.lists {
		log.Debugf("Loading list %v", list)

		var src io.ReadCloser
		if strings.HasPrefix(list, "http") {
			resp, err := http.Get(list)
			if err != nil {
				return err
			}
			src = resp.Body

		} else if strings.HasPrefix(list, ".") {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			file, err := os.Open(cwd + strings.TrimPrefix(list, "."))
			if err != nil {
				return err
			}
			src = file

		} else {
			file, err := os.Open(list)
			if err != nil {
				return err
			}
			src = file
		}

		if block {
			if err := blocklist.Read(src); err != nil {
				return err
			}
		} else {
			if err := whitelist.Read(src); err != nil {
				return err
			}
		}
	}

	f.mu.Lock()
	f.whitelist = whitelist
	f.blacklist = blocklist
	f.mu.Unlock()

	return nil
}

func (f *Filter) Reload() error {
	return f.Load()
}
