package filter

import (
	"context"
	"net"
	"strings"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

// Filter represents a plugin instance that can filter and block requests based
// on predefined lists and regex rules.
type Filter struct {
	Next plugin.Handler

	// Enables CNAME uncloaking of replies.
	UncloakCname bool

	sources   []source
	whitelist *dnsFilter
	blacklist *dnsFilter
}

func New() *Filter {
	return &Filter{
		whitelist: newDNSFilter(),
		blacklist: newDNSFilter(),
	}
}

// ServeDNS implements the plugin.Handler interface.
func (f *Filter) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	server := metrics.WithServer(ctx)
	qname := state.Name()
	answers := []dns.RR{}

	if f.Match(qname) {
		BlockCount.WithLabelValues(server).Inc()

		switch state.QType() {
		case dns.TypeA:
			a := new(dns.A)
			a.Hdr = dns.RR_Header{
				Name:   qname,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    600,
			}
			a.A = net.IPv4zero
			answers = append(answers, a)

		case dns.TypeAAAA:
			a := new(dns.AAAA)
			a.Hdr = dns.RR_Header{
				Name:   qname,
				Rrtype: dns.TypeAAAA,
				Class:  dns.ClassINET,
				Ttl:    600,
			}
			a.AAAA = net.IPv6zero
			answers = append(answers, a)

		default:
			return plugin.NextOrFailure(f.Name(), f.Next, ctx, w, r)
		}

		m := new(dns.Msg)
		m.SetReply(r)
		m.SetRcode(r, dns.RcodeSuccess)
		m.Authoritative = true
		m.Answer = answers

		w.WriteMsg(m)
		return dns.RcodeSuccess, nil
	}

	if f.UncloakCname {
		fw := &ResponseWriter{ResponseWriter: w, state: state, server: server, Filter: f}
		return plugin.NextOrFailure(f.Name(), f.Next, ctx, fw, r)
	}

	return plugin.NextOrFailure(f.Name(), f.Next, ctx, w, r)
}

// Name implements the plugin.Handler interface.
func (f *Filter) Name() string {
	return "filter"
}

func (f *Filter) OnStartup() error {
	return f.Load()
}

func (f *Filter) Match(name string) bool {
	if f.whitelist.Match(name) {
		return false
	}
	if f.blacklist.Match(name) {
		return true
	}
	return false
}

func (f *Filter) Load() error {
	for _, list := range f.sources {
		rc, err := list.Read()
		if err != nil {
			return err
		}
		if list.Block {
			if _, err := f.blacklist.ReadFrom(rc); err != nil {
				return err
			}
		} else {
			if _, err := f.whitelist.ReadFrom(rc); err != nil {
				return err
			}
		}
		rc.Close()
	}

	return nil
}

// ResponseWriter is a response writer that performs CNAME uncloaking.
type ResponseWriter struct {
	dns.ResponseWriter
	*Filter

	state  request.Request
	server string
}

// WriteMsg implements the dns.ResponseWriter interface.
func (w *ResponseWriter) WriteMsg(m *dns.Msg) error {
	if m.Rcode != dns.RcodeSuccess {
		return w.ResponseWriter.WriteMsg(m)
	}

	if w.whitelist.Match(w.state.Name()) {
		return w.ResponseWriter.WriteMsg(m)
	}

	for _, r := range m.Answer {
		hdr := r.Header()
		if hdr.Class != dns.ClassINET || hdr.Rrtype != dns.TypeCNAME {
			continue
		}

		cname := strings.TrimSuffix(r.(*dns.CNAME).Target, ".")
		if w.Match(cname) {
			BlockCount.WithLabelValues(w.server).Inc()

			r := w.state.Req
			msg := new(dns.Msg)
			msg.SetReply(r)
			msg.SetRcode(r, dns.RcodeNameError)

			w.WriteMsg(msg)
			return nil
		}
	}
	return w.ResponseWriter.WriteMsg(m)
}

// Write implements the dns.ResponseWriter interface.
func (w *ResponseWriter) Write(buf []byte) (int, error) {
	// log ?
	return w.ResponseWriter.Write(buf)
}
