package filter

import (
	"testing"

	"github.com/caddyserver/caddy"
)

func TestSetup(t *testing.T) {
	c := caddy.NewTestController("dns", `filter`)
	if err := setup(c); err == nil {
		t.Errorf("Expected errors, but got: %v", err)
	}

	c = caddy.NewTestController("dns", `filter {
		allow https://dl.paesacybersecurity.eu/lists/whitelist.txt
		block https://dl.paesacybersecurity.eu/lists/blacklist.txt
	}`)
	if err := setup(c); err != nil {
		t.Errorf("Expected no errors, but got: %v", err)
	}
}
