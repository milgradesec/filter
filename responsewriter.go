package filter

import (
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

type ResponseWriter struct {
	dns.ResponseWriter
	*Filter

	state request.Request
}

func (w *ResponseWriter) WriteMsg(m *dns.Msg) error {
	for _, r := range m.Answer {
		hdr := r.Header()
		if hdr.Class != dns.ClassINET || hdr.Rrtype != dns.TypeCNAME {
			continue
		}

		/*cname := r.(*dns.CNAME).Target
		if w.Plugin.Query(cname, false) {
			log.Infof("Blocked CNAME")

			resp := new(dns.Msg)
			resp.SetRcode(w.Request, dns.RcodeNameError)
			return w.WriteMsg(resp)
		}*/
	}
	return w.ResponseWriter.WriteMsg(m)
}

func (w *ResponseWriter) Write(buf []byte) (int, error) {
	log.Warning("Filter called with Write: not filtering reply")

	return w.ResponseWriter.Write(buf)
}
