package filter

import (
	"testing"

	"github.com/caddyserver/caddy"
)

func TestSetup(t *testing.T) {
	c := caddy.NewTestController("dns", `filter {
		list /lists/whitelist.txt white 
		list /lists/blacklist.txt black 
		list /lists/privatelist.txt private
        list /lists/blocklist.txt black
		}`)
	if err := setup(c); err != nil {
		t.Errorf("Expected no errors, but got: %q", err)
	}
}
