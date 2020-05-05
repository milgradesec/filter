package filter

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"strings"
)

type patternMatcher struct {
	hashtable  map[string]struct{}
	prefixes   []string
	suffixes   []string
	subStrings []string
	regexes    []*regexp.Regexp
}

func newPatternMatcher() *patternMatcher {
	return &patternMatcher{
		hashtable: make(map[string]struct{}),
	}
}

var regexpRunes = []string{"[", "]", "(", ")", "|", "?",
	"+", "$", "{", "}", "^"}

func (pm *patternMatcher) ReadFrom(r io.Reader) (n int64, err error) {
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
				pm.regexes = append(pm.regexes, r)
				break
			}
		}
		if strings.Contains(line, "*") {
			if strings.HasSuffix(line, "*") && strings.HasPrefix(line, "*") {
				qname := strings.TrimPrefix(line, "*")
				qname = strings.TrimSuffix(qname, "*")
				pm.subStrings = append(pm.subStrings, qname)
			}
			if strings.HasSuffix(scanner.Text(), "*") {
				domain := strings.TrimSuffix(line, "*")
				pm.prefixes = append(pm.prefixes, domain)
			}
			if strings.HasPrefix(scanner.Text(), "*") {
				domain := strings.TrimPrefix(line, "*")
				pm.suffixes = append(pm.suffixes, domain)
			}
		} else {
			pm.hashtable[line] = struct{}{}
		}

		if scanner.Err() != nil {
			return 0, scanner.Err()
		}
	}
	return n, nil
}

func (pm *patternMatcher) Match(str string) bool {
	_, q := pm.hashtable[str]
	if q {
		return true
	}
	for _, prefix := range pm.prefixes {
		if strings.HasPrefix(str, prefix) {
			return true
		}
	}
	for _, suffix := range pm.suffixes {
		if strings.HasSuffix(str, suffix) {
			return true
		}
		if str == strings.TrimPrefix(suffix, ".") {
			return true
		}
	}
	for _, substr := range pm.subStrings {
		if strings.Contains(str, substr) {
			return true
		}
	}
	for _, regex := range pm.regexes {
		if regex.MatchString(str) {
			return true
		}
	}
	return false
}
