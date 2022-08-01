package filter

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"

	iradix "github.com/hashicorp/go-immutable-radix"
)

type PatternMatcher struct {
	prefixes     *iradix.Tree
	suffixes     *iradix.Tree
	subStrings   []string
	regexes      []*regexp.Regexp
	exactStrings map[string]struct{}
}

func NewPatternMatcher() *PatternMatcher {
	return &PatternMatcher{
		prefixes:     iradix.New(),
		suffixes:     iradix.New(),
		exactStrings: map[string]struct{}{},
	}
}

func (pm *PatternMatcher) Add(pattern string) error {
	pattern = strings.TrimSpace(pattern)

	if pattern == "" || strings.HasPrefix(pattern, "#") {
		return nil
	}
	if strings.Contains(pattern, "#") {
		i := strings.Index(pattern, "#")
		pattern = strings.TrimSpace(pattern[:i])
	}
	for _, op := range regexpRunes {
		if strings.Contains(pattern, op) {
			r, err := regexp.Compile(pattern)
			if err != nil {
				return err
			}
			pm.regexes = append(pm.regexes, r)
			break
		}
	}
	if strings.Contains(pattern, "*") {
		if strings.HasSuffix(pattern, "*") && strings.HasPrefix(pattern, "*") { //nolint
			qname := strings.TrimPrefix(pattern, "*")
			qname = strings.TrimSuffix(qname, "*")
			pm.subStrings = append(pm.subStrings, qname)
		} else if strings.HasSuffix(pattern, "*") {
			domain := strings.TrimSuffix(pattern, "*")
			pm.prefixes, _, _ = pm.prefixes.Insert([]byte(domain), 1)
		} else if strings.HasPrefix(pattern, "*") {
			domain := strings.TrimPrefix(pattern, "*")
			pm.suffixes, _, _ = pm.suffixes.Insert([]byte(stringReverse(domain)), 1)
		}
	} else {
		pm.exactStrings[pattern] = struct{}{}
	}
	return nil
}

func (pm *PatternMatcher) Match(qname string) bool {
	qname = strings.TrimSuffix(qname, ".")

	_, found := pm.exactStrings[qname]
	if found {
		return true
	}
	_, _, found = pm.prefixes.Root().LongestPrefix([]byte(qname))
	if found {
		return true
	}
	_, _, found = pm.suffixes.Root().LongestPrefix([]byte(stringReverse(qname)))
	if found {
		return true
	}
	_, found = pm.suffixes.Root().Get([]byte(stringReverse(qname) + "."))
	if found {
		return true
	}
	for _, substr := range pm.subStrings {
		if strings.Contains(qname, substr) {
			return true
		}
	}
	for _, regex := range pm.regexes {
		if regex.MatchString(qname) {
			return true
		}
	}
	return false
}

func (pm *PatternMatcher) LoadRules(r io.Reader) error {
	if r == nil {
		return errors.New("invalid list source")
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		err := pm.Add(scanner.Text())
		if err != nil {
			log.Error(err)
		}
		if scanner.Err() != nil {
			return scanner.Err()
		}
	}
	return nil
}

var regexpRunes = []string{"[", "]", "(", ")", "|", "?",
	"+", "$", "{", "}", "^"}

func stringReverse(s string) string {
	size := len(s)
	buf := make([]byte, size)
	for start := 0; start < size; {
		r, n := utf8.DecodeRuneInString(s[start:])
		start += n
		utf8.EncodeRune(buf[size-start:], r)
	}
	return string(buf)
}
