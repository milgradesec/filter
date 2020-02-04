package filter

import (
	"context"
	"time"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

var log = clog.NewWithPlugin("filter")

type filter struct {
	Next plugin.Handler

	Lists          []*List
	BlockedTtl     uint32
	ReloadInterval time.Duration

	whitelist *PatternMatcher
	blacklist *PatternMatcher
}

func (f *filter) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	name := trimTrailingDot(state.Name())

	if f.Match(name) {
		BlockCount.WithLabelValues(metrics.WithServer(ctx)).Inc()
		return writeNXdomain(w, r)
	}

	rw := &ResponseWriter{
		ResponseWriter: w,
		filter:         f,
		state:          state,
	}
	return plugin.NextOrFailure(f.Name(), f.Next, ctx, rw, r)
}

func (f *filter) Match(str string) bool {
	if f.whitelist.Match(str) {
		return false
	}
	if f.blacklist.Match(str) {
		return true
	}
	return false
}

func (f *filter) Load() error {
	whitelist := NewPatternMatcher()
	blocklist := NewPatternMatcher()

	for _, list := range f.Lists {
		rc, err := list.Open()
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

func (f *filter) OnStartup() error {
	return f.Load()
}

func (f *filter) Name() string {
	return "filter"
}
