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
	return plugin.NextOrFailure(f.Name(), f.Next, ctx, w, r)
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

func (f *filter) OnStartup() error {
	return f.Load()
}

func (f *filter) OnShutdown() error {
	return nil
}

func (f *filter) Name() string {
	return "filter"
}
