package filter

import (
	"context"
	"testing"

	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"

	"github.com/miekg/dns"
)

func Test_ServeDNS(t *testing.T) {
	c := caddy.NewTestController("dns", `filter  {
		allow https://dl.paesacybersecurity.eu/lists/whitelist.txt
		block https://dl.paesacybersecurity.eu/lists/blacklist.txt
	}`)

	f, err := parseConfig(c)
	if err != nil {
		t.Fatal(err)
	}
	f.Next = test.NextHandler(dns.RcodeSuccess, nil)

	if err = f.Load(); err != nil {
		t.Fatal(err)
	}
	rec := dnstest.NewRecorder(&test.ResponseWriter{})

	tests := []struct {
		name  string
		block bool
	}{
		{"example.com", false},
		{"instagram.com", false},
		{"facebook.com", false},
		{"adservice.google.com", true},
		{"ads.example.com", true},
		{"mipcwtf.lan", true},
		{"example.taboola.com", true},
		{"beacons7.gvt2.com", true},
		{".", false},
	}

	for i, tt := range tests {
		req := new(dns.Msg)
		req.SetQuestion(tt.name, dns.TypeA)

		rcode, err := f.ServeDNS(context.TODO(), rec, req)
		if err != nil {
			t.Fatal(err)
		}
		if rcode == dns.RcodeNameError && !tt.block {
			t.Errorf("Test %d: expected NOERROR but got %s", i, dns.RcodeToString[rcode])
		}
		if rcode != dns.RcodeSuccess && rcode != dns.RcodeNameError {
			t.Errorf("Test %d: expected other rcode but got %s", i, dns.RcodeToString[rcode])
		}
	}
}
