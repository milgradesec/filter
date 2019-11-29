package filter

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

type filterType int

const (
	white filterType = iota
	black
	private
)

type filter struct {
	domains    map[string]struct{}
	prefixes   []string
	suffixes   []string
	subStrings []string
	regexes    []*regexp.Regexp

	Type filterType
	Path string
}

func NewFilter(path string) (*filter, error) {
	f := &filter{
		domains: make(map[string]struct{}),
		Type:    black,
		Path:    path,
	}

	err := f.Load()
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (f *filter) Query(domain string) bool {
	_, q := f.domains[domain]
	if q {
		return true
	}
	for _, prefix := range f.prefixes {
		if strings.HasPrefix(domain, prefix) {
			return true
		}
	}
	for _, suffix := range f.suffixes {
		if strings.HasSuffix(domain, suffix) {
			return true
		}
		if domain == strings.TrimPrefix(suffix, ".") {
			return true
		}
	}
	for _, substr := range f.subStrings {
		if strings.Contains(domain, substr) {
			return true
		}
	}
	for _, regex := range f.regexes {
		if regex.MatchString(domain) {
			return true
		}
	}
	return false
}

func (f *filter) Load() error {
	file, err := os.Open(f.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	var regexpOps = []string{"[", "]", "(", ")", "|", "?", "+", "$", "{", "}", "^"}

	scanner := bufio.NewScanner(file)
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
				f.regexes = append(f.regexes, r)
				break
			}
		}
		if strings.Contains(line, "*") {
			if strings.HasSuffix(line, "*") && strings.HasPrefix(line, "*") {
				domain := strings.TrimPrefix(line, "*")
				domain = strings.TrimSuffix(domain, "*")
				f.subStrings = append(f.subStrings, domain)
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
			f.domains[line] = struct{}{}
		}
	}
	return nil
}
