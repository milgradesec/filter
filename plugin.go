package filter

import (
	"strings"
	"sync"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

var log = clog.NewWithPlugin("filter")

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

		m := new(dns.Msg)
		m.SetRcode(r, dns.RcodeNameError)
		err := w.WriteMsg(m)
		if err != nil {
			return dns.RcodeServerFailure, err
		}
		return dns.RcodeNameError, nil
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
		log.Infof("Visto CNAME")

		if w.Plugin.Query(cname, false) {
			log.Infof("Blocked CNAME")

			resp := new(dns.Msg)
			resp.SetRcode(w.Request, dns.RcodeNameError)
			return w.WriteMsg(resp)
		}
	}
	return w.ResponseWriter.WriteMsg(m)
}

/*func replyBlockedResponse(w dns.ResponseWriter, r *dns.Msg) error {
	m := new(dns.Msg)
	m.SetReply(r)
	m.SetRcode(r, dns.RcodeNameError)
	hdr := dns.RR_Header{Name: r.Question[0].Name, Ttl: 60, Rrtype: dns.TypeSOA}
	//m.Answer = []dns.RR{&dns.HINFO{Hdr: hdr, Cpu: "BLOCKED"}}
	m.Answer = []dns.RR{&dns.SOA{Hdr: hdr, Minttl: 60}}
	return w.WriteMsg(m)
}*/
