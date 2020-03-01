package filter

import (
	"github.com/miekg/dns"
)

func genSOA(r *dns.Msg) []dns.RR {
	zone := r.Question[0].Name
	hdr := dns.RR_Header{
		Name:   zone,
		Rrtype: dns.TypeSOA,
		Ttl:    600,
		Class:  dns.ClassINET,
	}

	mbox := "hostmaster."
	if zone[0] != '.' {
		mbox += zone
	}

	soa := dns.SOA{
		Hdr:     hdr,
		Ns:      "fake-for-negative-caching.",
		Mbox:    mbox,
		Serial:  100500, // faster than uint32(time.Now().Unix())
		Refresh: 1800,
		Retry:   900,
		Expire:  604800,
		Minttl:  86400,
	}
	return []dns.RR{&soa}
}

func writeNXdomain(w dns.ResponseWriter, r *dns.Msg) (int, error) {
	m := new(dns.Msg)
	m.SetRcode(r, dns.RcodeNameError)
	m.Authoritative, m.RecursionAvailable = true, true
	m.Ns = genSOA(r)

	err := w.WriteMsg(m)
	if err != nil {
		return dns.RcodeServerFailure, err
	}
	return dns.RcodeNameError, nil
}
