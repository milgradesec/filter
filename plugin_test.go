package filter

import (
	"context"
	"testing"

	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"

	"github.com/miekg/dns"
)

func Test_Filter_Query(t *testing.T) {
	c := caddy.NewTestController("dns", `filter {
		list /lists/whitelist.txt white 
		list /lists/blacklist.txt black 
		list /lists/privatelist.txt private
        list /lists/blocklist.txt black
		}`)
	p, err := parseFilter(c)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		qname string
		block bool
	}{
		{
			qname: "example.com",
			block: false,
		},
		{
			qname: "google.com",
			block: false,
		},
		{
			qname: "instagram.com",
			block: false,
		},
		{
			qname: "facebook.com",
			block: false,
		},
		{
			qname: "adservice.google.com",
			block: true,
		},
		{
			qname: "ads.example.com",
			block: true,
		},
		{
			qname: "mipcwtf.lan",
			block: true,
		},
		{
			qname: "openload.co",
			block: false,
		},
		{
			qname: "www.rapidvideo.com",
			block: false,
		},
		{
			qname: "example.taboola.com",
			block: true,
		},
		{
			qname: "beacons.gvt1.com",
			block: true,
		},
		{
			qname: "beacons5.gvt2.com",
			block: true,
		},
		{
			qname: ".",
			block: false,
		},
	}
	for _, tt := range tests {
		if got := p.Query(tt.qname, false); got != tt.block {
			t.Errorf("Filter.Query(%s) = %v, want %v", tt.qname, got, tt.block)
		}
	}

	if !p.Query("facebook.com", true) {
		t.Errorf("Facebook not blocked")
	}

}

func Test_ServeDNS(t *testing.T) {
	c := caddy.NewTestController("dns", `filter {
		list /lists/whitelist.txt white 
		list /lists/blacklist.txt black 
		list /lists/privatelist.txt private
        list /lists/blocklist.txt black
		}`)
	f, err := parseFilter(c)
	if err != nil {
		t.Fatal(err)
	}
	f.Next = test.ErrorHandler()

	rec := dnstest.NewRecorder(&test.ResponseWriter{})
	req := new(dns.Msg)
	req.SetQuestion("example.org", dns.TypeA)

	_, err = f.ServeDNS(context.TODO(), rec, req)
	if err != nil {
		t.Error(err)
	}
}
