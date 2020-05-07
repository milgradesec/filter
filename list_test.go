package filter

import (
	"testing"

	"github.com/caddyserver/caddy"
)

func TestFilter_Load(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{`filter {
			allow ./tests/data/whitelist.txt
			block ./tests/data/blacklist.txt
		}`, false},
	}
	for _, tt := range tests {
		c := caddy.NewTestController("dns", tt.input)
		f, err := parseFilter(c)
		if err != nil {
			t.Fatal(err)
		}

		if err := f.Load(); (err != nil) != tt.wantErr {
			t.Errorf("Filter.Load() error = %v, wantErr %v", err, tt.wantErr)
		}
	}
}
