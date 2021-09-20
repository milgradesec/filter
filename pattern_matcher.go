package filter

import (
	"bufio"
	"errors"
	"io"
	"os"
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

func (f *PatternMatcher) Add(pattern string) error {
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
			f.regexes = append(f.regexes, r)
			break
		}
	}
	if strings.Contains(pattern, "*") {
		if strings.HasSuffix(pattern, "*") && strings.HasPrefix(pattern, "*") { //nolint
			qname := strings.TrimPrefix(pattern, "*")
			qname = strings.TrimSuffix(qname, "*")
			f.subStrings = append(f.subStrings, qname)
		} else if strings.HasSuffix(pattern, "*") {
			domain := strings.TrimSuffix(pattern, "*")
			f.prefixes, _, _ = f.prefixes.Insert([]byte(domain), 1)
		} else if strings.HasPrefix(pattern, "*") {
			domain := strings.TrimPrefix(pattern, "*")
			f.suffixes, _, _ = f.suffixes.Insert([]byte(stringReverse(domain)), 1)
		}
	} else {
		f.exactStrings[pattern] = struct{}{}
	}
	return nil
}

func (f *PatternMatcher) Match(qname string) bool {
	qname = strings.TrimSuffix(qname, ".")

	_, found := f.exactStrings[qname]
	if found {
		return true
	}
	_, _, found = f.prefixes.Root().LongestPrefix([]byte(qname))
	if found {
		return true
	}
	_, _, found = f.suffixes.Root().LongestPrefix([]byte(stringReverse(qname)))
	if found {
		return true
	}
	_, found = f.suffixes.Root().Get([]byte(stringReverse(qname) + "."))
	if found {
		return true
	}
	for _, substr := range f.subStrings {
		if strings.Contains(qname, substr) {
			return true
		}
	}
	for _, regex := range f.regexes {
		if regex.MatchString(qname) {
			return true
		}
	}
	return false
}

func (f *PatternMatcher) Load(r io.Reader) error {
	if r == nil {
		return errors.New("invalid list source")
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		err := f.Add(scanner.Text())
		if err != nil {
			log.Error(err)
		}
		if scanner.Err() != nil {
			return scanner.Err()
		}
	}
	return nil
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
