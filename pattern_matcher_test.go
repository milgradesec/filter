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

func TestLoadRulesReturnsScannerError(t *testing.T) {
	pm := NewPatternMatcher()
	line := strings.Repeat("a", bufio.MaxScanTokenSize+1)

	err := pm.LoadRules(strings.NewReader(line))
	if err == nil {
		t.Fatal("expected scanner error for oversized rule, got nil")
	}
}
