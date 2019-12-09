package filter

import (
	"testing"

	"github.com/caddyserver/caddy"
)

func TestSetup(t *testing.T) {
	tests := []struct {
		input string
		err   bool
	}{
		{`filter`, true},
		{`filter more`, true},
		{`filter {
			more
			}`, true},
		{`filter {
			block url more
			}`, true},
		{`filter {
			allow path more
			}`, true},
		{`filter {
			allow https://dl.paesacybersecurity.eu/lists/whitelist.txt
			block https://dl.paesacybersecurity.eu/lists/blacklist.txt
		}`, false},
	}

	for i, test := range tests {
		c := caddy.NewTestController("dns", test.input)
		err := setup(c)

		if test.err && err == nil {
			t.Errorf("Test %d: expected error but found %s for input %s", i, err, test.input)
		}

		if !test.err && err != nil {
			t.Errorf("Test %d: expected no error but found one for input %s, got: %v", i, test.input, err)
		}
	}
}
