package filter

import "github.com/miekg/dns"

var defaultSOA = &dns.SOA{
	// values copied from verisign's nonexistent .com domain
	// their exact values are not important in our use case because
	// they are used for domain transfers between  DNS servers
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
