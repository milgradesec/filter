package filter

import (
	"strings"
	"sync"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

//var log = clog.NewWithPlugin("filter")

type Plugin struct {
	Next plugin.Handler

	filters      []*filter
	sync.RWMutex // reload mutex
}

func (p *Plugin) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	qname := strings.ToLower(r.Question[0].Name)
	if qname == "." {
		return plugin.NextOrFailure(p.Name(), p.Next, ctx, w, r)
	}

	qname = strings.TrimSuffix(qname, ".")
	var local bool
	if strings.HasPrefix(state.IP(), "192.168.") {
		local = true
	}

	p.RLock()
	defer p.RUnlock()

	if p.Query(qname, local) {
		server := metrics.WithServer(ctx)
		blockedCount.WithLabelValues(server).Inc()
		return writeNXdomain(w, r)
	}

	//rw := &ResponseWriter{ResponseWriter: w, Plugin: p, Request: r}
	return plugin.NextOrFailure(p.Name(), p.Next, ctx, w, r)
}

func (p *Plugin) Query(domain string, local bool) bool {
	for _, list := range p.filters {
		if list.Type == private && !local {
			continue
		}

		if list.Query(domain) {
			return !(list.Type == white)
		}
	}
	return false
}

func (p *Plugin) Name() string {
	return "filter"
}

type ResponseWriter struct {
	dns.ResponseWriter
	*Plugin

	Request *dns.Msg
}

func (w *ResponseWriter) WriteMsg(m *dns.Msg) error {
	for _, r := range m.Answer {
		hdr := r.Header()
		if hdr.Class != dns.ClassINET || hdr.Rrtype != dns.TypeCNAME {
			continue
		}
		cname := r.(*dns.CNAME).Target

		if w.Plugin.Query(cname, false) {
			resp := new(dns.Msg)
			resp.SetRcode(w.Request, dns.RcodeNameError)
			return w.WriteMsg(resp)
		}
	}
	return w.ResponseWriter.WriteMsg(m)
}

func writeNXdomain(w dns.ResponseWriter, r *dns.Msg) (int, error) {
	m := new(dns.Msg)
	m.SetRcode(r, dns.RcodeNameError)
	m.Authoritative, m.RecursionAvailable, m.Compress = true, true, true
	m.Ns = genSOA(r)

	err := w.WriteMsg(m)
	if err != nil {
		return dns.RcodeServerFailure, err
	}
	return dns.RcodeNameError, nil
}

var defaultSOA = &dns.SOA{
	// values copied from verisign's nonexistent .com domain
	// their exact values are not important in our use case because
	// they are used for domain transfers between primary/secondary DNS servers
	Refresh: 1800,
	Retry:   900,
	Expire:  604800,
	Minttl:  86400,
}

func genSOA(r *dns.Msg) []dns.RR {
	zone := r.Question[0].Name
	hdr := dns.RR_Header{Name: zone, Rrtype: dns.TypeSOA, Ttl: 120, Class: dns.ClassINET}

	Mbox := "hostmaster."
	if zone[0] != '.' {
		Mbox += zone
	}

	soa := *defaultSOA
	soa.Hdr = hdr
	soa.Mbox = Mbox
	soa.Ns = "fake-for-negative-caching.paesacybersecurity.eu."
	soa.Serial = 100500 // faster than uint32(time.Now().Unix())
	return []dns.RR{&soa}
}
