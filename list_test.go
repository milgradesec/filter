package filter

import (
	"github.com/caddyserver/caddy"
	"runtime"
	"testing"
)

func TestFilter_Load(t *testing.T) {
	var input string
	if runtime.GOOS == "windows" {
		input = `filter {
			allow D:\repos\filter\lists\whitelist.txt
			block D:\repos\filter\lists\blacklist.txt
		}`
	} else {
		input = `filter {
			allow /home/travis/gopath/src/github.com/milgradesec/filter/lists/whitelist.txt
			block /home/travis/gopath/src/github.com/milgradesec/filter/lists/blacklist.txt
		}`
	}

	tests := []struct {
		input   string
		wantErr bool
	}{
		{`filter {
			allow ./lists/whitelist.txt
			block ./lists/blacklist.txt
		}`, false},
		{input, false},
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
