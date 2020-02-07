package filter

import (
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

type ResponseWriter struct {
	dns.ResponseWriter
	*filter

	state request.Request
}

func (w *ResponseWriter) WriteMsg(m *dns.Msg) error {
	qname := trimTrailingDot(w.state.Name())
	if m.Rcode != dns.RcodeSuccess || w.whitelist.Match(qname) {
		return w.ResponseWriter.WriteMsg(m)
	}

	for _, r := range m.Answer {
		hdr := r.Header()
		if hdr.Class != dns.ClassINET || hdr.Rrtype != dns.TypeCNAME {
			continue
		}

		cname := trimTrailingDot(r.(*dns.CNAME).Target)
		if w.Match(cname) {
			if _, err := writeNXdomain(w, w.state.Req); err != nil {
				return err
			}
			return nil
		}
	}
	return w.ResponseWriter.WriteMsg(m)
}
