package filter

import (
	"testing"

	"github.com/caddyserver/caddy"
)

func TestSetup(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{`filter`, true},
		{`filter more`, true},
		{`filter {
			more
			}`, true},
		{`filter {
			block url more
			allow path more
			}`, true},
		{`filter {
			block 
			allow
			ttl
			}`, true},
		{`filter {
			allow https://dl.paesacybersecurity.eu/lists/whitelist.txt
			block https://dl.paesacybersecurity.eu/lists/blacklist.txt
		}`, false},
		{`filter {
			allow ./lists/whitelist.txt
			block ./lists/blacklist.txt
			ttl 600
		}`, false},
	}

	for i, test := range tests {
		c := caddy.NewTestController("dns", test.input)
		err := setup(c)

		if test.wantErr && err == nil {
			t.Errorf("Test %d: expected error but found %s for input %s", i, err, test.input)
		}
		if !test.wantErr && err != nil {
			t.Errorf("Test %d: expected no error but found one for input %s, got: %v", i, test.input, err)
		}
	}
}
