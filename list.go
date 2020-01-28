package filter

import (
	"bufio"
	"errors"
	"io"
	"os"
	"regexp"
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

type PatternMatcher struct {
	hashtable  map[string]struct{}
	prefixes   []string
	suffixes   []string
	subStrings []string
	regexes    []*regexp.Regexp
}

func NewPatternMatcher() *PatternMatcher {
	return &PatternMatcher{
		hashtable: make(map[string]struct{}),
	}
}

var regexpRunes = []string{"[", "]", "(", ")", "|", "?",
	"+", "$", "{", "}", "^"}

func (l *PatternMatcher) ReadFrom(r io.Reader) (n int64, err error) {
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
			return 0, scanner.Err()
		}
	}
	return n, nil
}

func (l *PatternMatcher) Match(str string) bool {
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
