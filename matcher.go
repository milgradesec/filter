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

type matcher struct {
	hashtable  map[string]struct{}
	prefixes   *iradix.Tree
	suffixes   *iradix.Tree
	subStrings []string
	regexes    []*regexp.Regexp
}

func newMatcher() *matcher {
	return &matcher{
		hashtable: map[string]struct{}{},
		prefixes:  iradix.New(),
		suffixes:  iradix.New(),
	}
}

func (f *matcher) Match(name string) bool {
	name = strings.TrimSuffix(name, ".")

	_, found := f.hashtable[name]
	if found {
		return true
	}
	_, _, found = f.prefixes.Root().LongestPrefix([]byte(name))
	if found {
		return true
	}
	_, _, found = f.suffixes.Root().LongestPrefix([]byte(stringReverse(name)))
	if found {
		return true
	}
	_, found = f.suffixes.Root().Get([]byte(stringReverse(name) + "."))
	if found {
		return true
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

func (f *matcher) Load(r io.Reader) (n int64, err error) {
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
		if strings.Contains(line, "#") {
			i := strings.Index(line, "#")
			line = strings.TrimSpace(line[:i])
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
				f.prefixes, _, _ = f.prefixes.Insert([]byte(domain), 1)
			}
			if strings.HasPrefix(scanner.Text(), "*") {
				domain := strings.TrimPrefix(line, "*")
				f.suffixes, _, _ = f.suffixes.Insert([]byte(stringReverse(domain)), 1)
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
