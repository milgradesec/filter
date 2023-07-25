package filter

import (
	"bytes"
	"context"
	"crypto/sha256"
	"io"
	"strings"
	"time"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	ranger "github.com/yl2chen/cidranger"
)

const defaultResponseTTL = 3600 // Default TTL used for generated responses.

// Filter represents a plugin instance that can filter and block requests based
// on predefined lists and regex rules.
type Filter struct {
	Next plugin.Handler

	// lists to allow or block domains from a file
	allowlist *PatternMatcher
	denylist  *PatternMatcher

	// lists to allow or block records
	allowCIDR4list ranger.Ranger
	allowCIDR6list ranger.Ranger
	denyCIDR4list  ranger.Ranger
	denyCIDR6list  ranger.Ranger

	reload time.Duration
	hash   []byte

	// return empty answers in the requests.
	emptyResponse bool

	// sources to load data into filters.
	sources []listSource

	// uncloak enables response inspection for CNAME cloaking.
	uncloak bool

	// ttl sets a custom TTL.
	ttl uint32
}

func New() *Filter {
	return &Filter{
		allowlist:      NewPatternMatcher(),
		denylist:       NewPatternMatcher(),
		allowCIDR4list: ranger.NewPCTrieRanger(),
		allowCIDR6list: ranger.NewPCTrieRanger(),
		denyCIDR4list:  ranger.NewPCTrieRanger(),
		denyCIDR6list:  ranger.NewPCTrieRanger(),
		ttl:            defaultResponseTTL,
	}
}

// ServeDNS implements the plugin.Handler interface.
func (f *Filter) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	server := metrics.WithServer(ctx)

	if f.Match(state.Name()) {
		BlockCount.WithLabelValues(server).Inc()

		var msg *dns.Msg
		if !f.emptyResponse {
			msg = createSyntheticResponse(r, f.ttl)
		} else {
			msg = newEmptyResponse(r, f.ttl)
		}
		w.WriteMsg(msg) //nolint
		return dns.RcodeSuccess, nil
	}

	if f.uncloak {
		rw := &ResponseWriter{ResponseWriter: w, state: state, server: server, Filter: f}
		return plugin.NextOrFailure(f.Name(), f.Next, ctx, rw, r)
	}

	return plugin.NextOrFailure(f.Name(), f.Next, ctx, w, r)
}

// Name implements the plugin.Handler interface.
func (f *Filter) Name() string {
	return "filter"
}

// Match determines if the requested name should be blocked or allowed.
func (f *Filter) Match(name string) bool {
	if f.allowlist.Match(name) {
		return false
	}
	if f.denylist.Match(name) {
		return true
	}
	return false
}

// Does a hash on the list files to determine if anything has changed
func (f *Filter) checkHash() (ck bool) {
	h := sha256.New()
	for _, src := range f.sources {
		rc, err := src.Open()
		if err != nil {
			log.Error(err)
			return false
		}
		defer rc.Close()
		_, err = io.Copy(h, rc)
		if err != nil {
			log.Error(err)
			return false
		}
		rc.Close()
	}
	s := h.Sum(nil)
	ck = bytes.Compare(s, f.hash) != 0
	f.hash = s
	return
}

// Load in the files and set the denylist and allowlist if no errors are encountered
func (f *Filter) Load() error {
	denylist := NewPatternMatcher()
	allowlist := NewPatternMatcher()
	allowCIDR4list := ranger.NewPCTrieRanger()
	allowCIDR6list := ranger.NewPCTrieRanger()
	denyCIDR4list := ranger.NewPCTrieRanger()
	denyCIDR6list := ranger.NewPCTrieRanger()
	for _, src := range f.sources {
		rc, err := src.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		if !src.IsCIDR {
			if src.IsBlock {
				if err := denylist.LoadRules(rc); err != nil {
					return err
				}
			} else {
				if err := allowlist.LoadRules(rc); err != nil {
					return err
				}
			}
		} else {
			if src.IsBlock {
				if err := LoadCIDR(rc, denyCIDR4list, denyCIDR6list); err != nil {
					return err
				}
			} else {
				if err := LoadCIDR(rc, allowCIDR4list, allowCIDR6list); err != nil {
					return err
				}
			}
		}
		rc.Close()
	}
	f.denylist = denylist
	f.allowlist = allowlist
	f.allowCIDR4list = allowCIDR4list
	f.allowCIDR6list = allowCIDR6list
	f.denyCIDR4list = denyCIDR4list
	f.denyCIDR6list = denyCIDR6list

	return nil
}

// ResponseWriter is a response writer that performs response uncloaking.
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

	if w.allowlist.Match(w.state.Name()) {
		return w.ResponseWriter.WriteMsg(m)
	}

	var answers []dns.RR
	for _, r := range m.Answer {
		header := r.Header()
		if header.Class != dns.ClassINET {
			answers = append(answers, r)
			continue
		}

		var target string
		switch header.Rrtype {
		case dns.TypeCNAME:
			target = r.(*dns.CNAME).Target //nolint
		case dns.TypeSVCB:
			target = r.(*dns.SVCB).Target //nolint
		case dns.TypeHTTPS:
			target = r.(*dns.HTTPS).Target //nolint
		case dns.TypeA:
			ip := r.(*dns.A).A //nolint
			if c, err := w.denyCIDR4list.Contains(ip); !c && err == nil {
				answers = append(answers, r)
			} else if c, err := w.allowCIDR4list.Contains(ip); !c && err == nil {
				answers = append(answers, r)
			}
			continue
		case dns.TypeAAAA:
			ip := r.(*dns.AAAA).AAAA //nolint
			if c, err := w.denyCIDR6list.Contains(ip); !c && err == nil {
				answers = append(answers, r)
			} else if c, err := w.allowCIDR6list.Contains(ip); !c && err == nil {
				answers = append(answers, r)
			}
			continue
		default:
			answers = append(answers, r)
			continue
		}

		target = strings.TrimSuffix(target, ".")
		if w.Match(target) {
			BlockCount.WithLabelValues(w.server).Inc()

			r := w.state.Req
			var msg *dns.Msg
			if !w.emptyResponse {
				msg = createSyntheticResponse(r, w.ttl)
			} else {
				msg = newEmptyResponse(r, w.ttl)
			}
			w.WriteMsg(msg) //nolint
			return nil
		}
		answers = append(answers, r)
	}

	// If all the answers were stripped away, return server failure.  Doing so may make the client retry and get a new set of IPs.
	if len(m.Answer) > 0 && len(answers) == 0 {
		m.Rcode = dns.RcodeServerFailure
	}

	m.Answer = answers
	return w.ResponseWriter.WriteMsg(m)
}
