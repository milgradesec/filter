package filter

import (
	"testing"

	"github.com/caddyserver/caddy"
)

// Test loading lists

func TestFilter_Load(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{`filter {
			allow https://dl.paesacybersecurity.eu/lists/whitelist.txt
			block https://dl.paesacybersecurity.eu/lists/blacklist.txt
		}`, false},
		{`filter {
			allow ./lists/whitelist.txt
			block ./lists/blacklist.txt
		}`, false},
		{`filter {
			allow D:\repos\filter\lists\whitelist.txt
			block D:\repos\filter\lists\blacklist.txt
		}`, false},
	}
	for _, tt := range tests {
		c := caddy.NewTestController("dns", tt.input)
		f, err := parseConfig(c)
		if err != nil {
			t.Fatal(err)
		}

		if err := f.Load(); (err != nil) != tt.wantErr {
			t.Errorf("Filter.Load() error = %v, wantErr %v", err, tt.wantErr)
		}
	}
}
