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
// on predefined lists and regex rules.
type Filter struct {
	lists []*list

	whitelist    *PatternMatcher
	blacklist    *PatternMatcher
	uncloakCname bool

	Next plugin.Handler
}

func New() *Filter {
	return &Filter{
		whitelist: NewPatternMatcher(),
		blacklist: NewPatternMatcher(),
	}
}

// Name implements the plugin.Handler interface.
func (f *Filter) Name() string {
	return "filter"
}

// ServeDNS implements the plugin.Handler interface.
func (f *Filter) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	name := strings.TrimSuffix(state.Name(), ".")

	if f.Match(name) {
		BlockCount.WithLabelValues(metrics.WithServer(ctx)).Inc()

		m := new(dns.Msg)
		m.SetRcode(r, dns.RcodeNameError)
		m.Authoritative, m.RecursionAvailable = true, true
		m.Ns = genSOA(r)

		w.WriteMsg(m)
		return dns.RcodeNameError, nil
	}

	/*if f.uncloakCname {
		rw := &responseWriter{
			ResponseWriter: w,
			Filter:         f,
			state:          state,
		}
		return plugin.NextOrFailure(f.Name(), f.Next, ctx, rw, r)
	}*/

	return plugin.NextOrFailure(f.Name(), f.Next, ctx, w, r)
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
	for _, list := range f.lists {
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
