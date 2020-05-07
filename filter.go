package filter

import (
	"context"
	"strings"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

// Filter represents a plugin instance that can filter and block requests based
// on predefined lists.
type Filter struct {
	Next  plugin.Handler
	Lists []*list

	whitelist    *patternMatcher
	blacklist    *patternMatcher
	uncloakCname bool
}

// Name implements plugin.Handler.
func (f *Filter) Name() string {
	return "filter"
}

// ServeDNS implements plugin.Handler.
func (f *Filter) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	qname := strings.TrimSuffix(state.Name(), ".")

	if f.Match(qname) {
		BlockCount.WithLabelValues(metrics.WithServer(ctx)).Inc()
		return writeNXdomain(w, r)
	}

	if f.uncloakCname {
		rw := &responseWriter{
			ResponseWriter: w,
			Filter:         f,
			state:          state,
		}
		return plugin.NextOrFailure(f.Name(), f.Next, ctx, rw, r)
	}
	return plugin.NextOrFailure(f.Name(), f.Next, ctx, w, r)
}

// Match determines if the requested domain should be blocked.
func (f *Filter) Match(qname string) bool {
	if f.whitelist.Match(qname) {
		return false
	}
	if f.blacklist.Match(qname) {
		return true
	}
	return false
}

// Load loads the lists from disk.
func (f *Filter) Load() error {
	whitelist := newPatternMatcher()
	blocklist := newPatternMatcher()

	for _, list := range f.Lists {
		rc, err := list.Read()
		if err != nil {
			return err
		}
		if list.Block {
			if _, err := blocklist.ReadFrom(rc); err != nil {
				return err
			}
		} else {
			if _, err := whitelist.ReadFrom(rc); err != nil {
				return err
			}
		}
		rc.Close()
	}

	f.whitelist = whitelist
	f.blacklist = blocklist
	return nil
}
