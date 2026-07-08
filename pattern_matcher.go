package filter

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"strings"

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
			return nil
		}
	}
	if strings.Contains(pattern, "*") {
		if strings.HasSuffix(pattern, "*") && strings.HasPrefix(pattern, "*") { //nolint
			qname := strings.TrimSuffix(strings.TrimPrefix(pattern, "*"), "*")
			pm.subStrings = append(pm.subStrings, qname)
		} else if domain, ok := strings.CutSuffix(pattern, "*"); ok {
			pm.prefixes, _, _ = pm.prefixes.Insert([]byte(domain), 1)
		} else if domain, ok := strings.CutPrefix(pattern, "*"); ok {
			pm.suffixes, _, _ = pm.suffixes.Insert([]byte(stringReverse(domain)), 1)
		}
	} else {
		pm.exactStrings[pattern] = struct{}{}
	}
	return nil
}

func (pm *PatternMatcher) Match(qname string) bool {
	qname = strings.TrimSuffix(qname, ".")

	if _, found := pm.exactStrings[qname]; found {
		return true
	}
	if _, _, found := pm.prefixes.Root().LongestPrefix([]byte(qname)); found {
		return true
	}
	if pm.suffixes.Len() > 0 {
		rev := []byte(stringReverse(qname))
		if _, _, found := pm.suffixes.Root().LongestPrefix(rev); found {
			return true
		}
		if _, found := pm.suffixes.Root().Get(append(rev, '.')); found {
			return true
		}
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
	}
	return scanner.Err()
}

var regexpRunes = []string{"[", "]", "(", ")", "|", "?",
	"+", "$", "{", "}", "^"}

// stringReverse reverses s byte by byte. DNS names are byte sequences, so
// no UTF-8 decoding is needed; rune-based reversal panics or corrupts the
// result on non-UTF-8 input.
func stringReverse(s string) string {
	b := []byte(s)
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return string(b)
}
