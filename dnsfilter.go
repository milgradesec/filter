package filter

import (
	"bufio"
	"errors"
	"io"
	"os"
	"regexp"
	"strings"
)

type dnsFilter struct {
	hashtable  map[string]struct{}
	prefixes   []string
	suffixes   []string
	subStrings []string
	regexes    []*regexp.Regexp
}

func newDNSFilter() *dnsFilter {
	return &dnsFilter{
		hashtable: make(map[string]struct{}),
	}
}

var regexpRunes = []string{"[", "]", "(", ")", "|", "?",
	"+", "$", "{", "}", "^"}

func (f *dnsFilter) ReadFrom(r io.Reader) (n int64, err error) {
	if r == nil {
		return 0, errors.New("invalid list source")
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		for _, op := range regexpRunes {
			if strings.Contains(line, op) {
				r, err := regexp.Compile(line)
				if err != nil {
					return 0, err
				}
				f.regexes = append(f.regexes, r)
				break
			}
		}
		if strings.Contains(line, "*") {
			if strings.HasSuffix(line, "*") && strings.HasPrefix(line, "*") {
				qname := strings.TrimPrefix(line, "*")
				qname = strings.TrimSuffix(qname, "*")
				f.subStrings = append(f.subStrings, qname)
			}
			if strings.HasSuffix(scanner.Text(), "*") {
				domain := strings.TrimSuffix(line, "*")
				f.prefixes = append(f.prefixes, domain)
			}
			if strings.HasPrefix(scanner.Text(), "*") {
				domain := strings.TrimPrefix(line, "*")
				f.suffixes = append(f.suffixes, domain)
			}
		} else {
			f.hashtable[line] = struct{}{}
		}

		if scanner.Err() != nil {
			return 0, scanner.Err()
		}
	}
	return n, nil
}

func (f *dnsFilter) Match(name string) bool {
	name = strings.TrimSuffix(name, ".")

	_, q := f.hashtable[name]
	if q {
		return true
	}
	for _, prefix := range f.prefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	for _, suffix := range f.suffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
		if name == strings.TrimPrefix(suffix, ".") {
			return true
		}
	}
	for _, substr := range f.subStrings {
		if strings.Contains(name, substr) {
			return true
		}
	}
	for _, regex := range f.regexes {
		if regex.MatchString(name) {
			return true
		}
	}
	return false
}

type source struct {
	Path  string
	Block bool
}

func (s *source) Read() (src io.ReadCloser, err error) {
	f, err := os.Open(s.Path)
	if err != nil {
		return nil, err
	}
	return f, nil
}
