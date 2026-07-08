package filter

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

func TestLoadRulesFromFile(t *testing.T) {
	f, err := os.Open("testdata/denylist.list")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	pm := NewPatternMatcher()
	if err := pm.LoadRules(f); err != nil {
		t.Fatal(err)
	}

	tests := []string{
		"adservice.google.com.",
		"cdn.outbrain.com.",
		"example.tracker.global.",
		"malware.com.co.",
	}
	for _, tt := range tests {
		if !pm.Match(tt) {
			t.Errorf("expected %s to match denylist", tt)
		}
	}
}

func TestStringReverse(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"a", "a"},
		{"example.com", "moc.elpmaxe"},
		// Non-UTF-8 input must not panic or drop bytes.
		{"\xffads.com", "moc.sda\xff"},
		{"ads\xff.com", "moc.\xffsda"},
	}
	for _, tt := range tests {
		if got := stringReverse(tt.in); got != tt.want {
			t.Errorf("stringReverse(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestAddNonUTF8Pattern(t *testing.T) {
	pm := NewPatternMatcher()
	if err := pm.Add("*\xffads.com"); err != nil {
		t.Fatal(err)
	}
	if !pm.Match("foo\xffads.com") {
		t.Error("expected suffix rule with non-UTF-8 bytes to match")
	}
	if pm.Match("foo.ads.com") {
		t.Error("unexpected match for name without the non-UTF-8 byte")
	}
}

func TestLoadRulesReturnsScannerError(t *testing.T) {
	pm := NewPatternMatcher()
	line := strings.Repeat("a", bufio.MaxScanTokenSize+1)

	err := pm.LoadRules(strings.NewReader(line))
	if err == nil {
		t.Fatal("expected scanner error for oversized rule, got nil")
	}
}
