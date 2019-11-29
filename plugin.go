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

type Plugin struct {
	Next plugin.Handler

	filters      []*filter
	sync.RWMutex // reload mutex
}

func (p *Plugin) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	qname := strings.TrimSuffix(r.Question[0].Name, ".")

	var local bool
	if strings.HasPrefix(state.IP(), "192.168.") {
		local = true
	}

	p.RLock()
	defer p.RUnlock()

	if p.Query(qname, local) {
		resp := new(dns.Msg)
		resp.SetRcode(r, dns.RcodeNameError)
		err := w.WriteMsg(resp)
		if err != nil {
			return dns.RcodeServerFailure, err
		}

		server := metrics.WithServer(ctx)
		blockedCount.WithLabelValues(server).Inc()

		return dns.RcodeNameError, nil
	}
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

func (p *Plugin) Name() string { return "filter" }

/*func replyBlockedResponse(w dns.ResponseWriter, r *dns.Msg) error {
	m := new(dns.Msg)
	m.SetReply(r)
	hdr := dns.RR_Header{Name: r.Question[0].Name, Ttl: 60, Rrtype: dns.TypeHINFO}
	m.Answer = []dns.RR{&dns.HINFO{Hdr: hdr, Cpu: "BLOCKED"}}
	err := w.WriteMsg(m)
	if err != nil {
		return err
	}
	return nil
}*/
