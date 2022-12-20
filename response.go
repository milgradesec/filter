package filter

import (
	"net"

	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

func createSyntheticResponse(r *dns.Msg, ttl uint32) *dns.Msg {
	state := request.Request{Req: r}

	switch state.QType() {
	case dns.TypeA:
		return newAResponse(r, ttl)

	case dns.TypeAAAA:
		return newAAAAResponse(r, ttl)

	default:
		return newNXDomainResponse(r, ttl)
	}
}

func newAResponse(r *dns.Msg, ttl uint32) *dns.Msg {
	a := new(dns.A)
	a.Hdr = dns.RR_Header{
		Name:   r.Question[0].Name,
		Rrtype: dns.TypeA,
		Class:  dns.ClassINET,
		Ttl:    ttl,
	}
	a.A = net.IPv4zero

	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.SetRcode(r, dns.RcodeSuccess)
	msg.Answer = []dns.RR{a}
	msg.Authoritative = true
	msg.RecursionAvailable = true
	msg.Compress = true
	return msg
}

func newAAAAResponse(r *dns.Msg, ttl uint32) *dns.Msg {
	aaaa := new(dns.AAAA)
	aaaa.Hdr = dns.RR_Header{
		Name:   r.Question[0].Name,
		Rrtype: dns.TypeAAAA,
		Class:  dns.ClassINET,
		Ttl:    ttl,
	}
	aaaa.AAAA = net.IPv6zero

	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.SetRcode(r, dns.RcodeSuccess)
	msg.Answer = []dns.RR{aaaa}
	msg.Authoritative = true
	msg.RecursionAvailable = true
	msg.Compress = true
	return msg
}

func newNXDomainResponse(r *dns.Msg, ttl uint32) *dns.Msg {
	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.SetRcode(r, dns.RcodeNameError)

	msg.Ns = []dns.RR{&dns.SOA{
		Refresh: 1800,
		Retry:   900,
		Expire:  604800,
		Minttl:  86400,
		Ns:      "nobody.dns.paesa.es.",
		Serial:  100500,

		Hdr: dns.RR_Header{
			Name:   r.Question[0].Name,
			Rrtype: dns.TypeSOA,
			Ttl:    ttl,
			Class:  dns.ClassINET,
		},
		Mbox: "hostmaster.", // zone will be appended later if it's not empty or "."
	}}
	return msg
}
