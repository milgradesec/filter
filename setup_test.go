package filter

import (
	"testing"

	"github.com/coredns/caddy"
)

func TestSetup(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{`filter {}`, false},
		{`filter { 
			invalid
			}`, true},
		{`filter {
			allow ./testdata/allowlist.list
			block ./testdata/denylist.list
			ttl 300
		}`, false},
		{`filter {
			allow ./testdata/allowlist.list
			block ./testdata/denylist.list
			uncloak
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
