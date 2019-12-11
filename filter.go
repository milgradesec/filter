package filter

import (
	"context"
	"errors"
	"sync"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

var log = clog.NewWithPlugin("filter")

type Filter struct {
	Next plugin.Handler

	lists     map[string]bool
	mu        sync.RWMutex
	whitelist *list
	blacklist *list

	ttl uint32
}

func New() *Filter {
	return &Filter{
		lists: make(map[string]bool),
	}
}

func (f *Filter) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	if len(r.Question) != 1 {
		return dns.RcodeFormatError, errors.New("DNS request with multiple questions")
	}

	state := request.Request{W: w, Req: r}
	name := trimTrailingDot(state.Name())

	if f.Match(name) {
		BlockCount.WithLabelValues(metrics.WithServer(ctx)).Inc()
		return writeNXdomain(w, r)
	}

	rw := &ResponseWriter{ResponseWriter: w, Filter: f, state: state}
	return plugin.NextOrFailure(f.Name(), f.Next, ctx, rw, r)
}

func (f *Filter) Match(str string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.whitelist.Match(str) {
		return false
	}
	if f.blacklist.Match(str) {
		return true
	}
	return false
}

func (f *Filter) OnStartup() error {
	return f.Load()
}

func (f *Filter) OnShutdown() error {
	return nil
}

func (f *Filter) Name() string {
	return "filter"
}
