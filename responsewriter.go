package filter

import (
	"strings"

	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

type responseWriter struct {
	dns.ResponseWriter
	*Filter

	state request.Request
}

// WriteMsg implements dns.ResponseWriter.
func (w *responseWriter) WriteMsg(m *dns.Msg) error {
	name := strings.TrimSuffix(w.state.Name(), ".")

	if m.Rcode != dns.RcodeSuccess || w.whitelist.Match(name) {
		return w.ResponseWriter.WriteMsg(m)
	}

	for _, r := range m.Answer {
		hdr := r.Header()
		if hdr.Class != dns.ClassINET || hdr.Rrtype != dns.TypeCNAME {
			continue
		}

		cname := strings.TrimSuffix(r.(*dns.CNAME).Target, ".")
		if w.Match(cname) {
			if _, err := writeNXdomain(w, w.state.Req); err != nil {
				return err
			}
			return nil
		}
	}
	return w.ResponseWriter.WriteMsg(m)
}
