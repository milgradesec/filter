package filter

import (
	"context"
	"testing"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/coredns/coredns/request"

	"github.com/miekg/dns"
)

func Test_ServeDNS(t *testing.T) {
	c := caddy.NewTestController("dns", `filter  {
		allow ./testdata/allowlist.list
		block ./testdata/denylist.list
	}`)

	f, err := parseFilter(c)
	if err != nil {
		t.Fatal(err)
	}
	f.Next = test.NextHandler(dns.RcodeSuccess, nil)

	if err = f.Load(); err != nil {
		t.Fatal(err)
	}
	rec := dnstest.NewRecorder(&test.ResponseWriter{})

	tests := []struct {
		name    string
		block   bool
		dnsType uint16
	}{
		{"example.com.", false, dns.TypeA},
		{"facebook.com.", false, dns.TypeAAAA},
		{"adservice.google.com.", true, dns.TypeA},
		{"ads.example.com.", true, dns.TypeAAAA},
		{"mipcwtf.lan.", true, dns.TypeA},
		{"taboola.com.", true, dns.TypeAAAA},
		{"example.taboola.com.", true, dns.TypeA},
		{".", false, dns.TypeA},
	}

	for i, tt := range tests {
		req := new(dns.Msg)
		req.SetQuestion(tt.name, tt.dnsType)

		rcode, err := f.ServeDNS(context.TODO(), rec, req)
		if err != nil {
			t.Fatal(err)
		}
		if rcode != dns.RcodeSuccess && rcode != dns.RcodeNameError {
			t.Errorf("Test %d: expected other rcode but got %s", i, dns.RcodeToString[rcode])
		}
	}
}

func Test_Uncloak(t *testing.T) {
	c := caddy.NewTestController("dns", `filter  {
		allow ./testdata/allowlist.list
		block ./testdata/denylist.list
	}`)

	f, err := parseFilter(c)
	if err != nil {
		t.Fatal(err)
	}
	f.Next = test.NextHandler(dns.RcodeSuccess, nil)

	if err = f.Load(); err != nil {
		t.Fatal(err)
	}

	req := new(dns.Msg)
	req.SetQuestion("notracker.example.com.", dns.TypeCNAME)

	m := new(dns.Msg)
	m.SetReply(req)
	m.Response, m.RecursionAvailable = true, true
	m.Answer = []dns.RR{test.CNAME("notracker.example.com. 3600 IN CNAME ads.tracker.com.")}

	state := request.Request{W: &test.ResponseWriter{}, Req: req}
	rw := &ResponseWriter{
		ResponseWriter: &test.ResponseWriter{},
		state:          state,
		server:         "test",
		Filter:         f,
	}

	rw.WriteMsg(m) //nolint
}
