package filter

import (
	"context"
	"sync"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

var log = clog.NewWithPlugin("filter")

type Filter struct {
	Next plugin.Handler

	lists map[string]bool
	mu    sync.RWMutex

	ttl uint32
}

func New() *Filter {
	return &Filter{
		lists: make(map[string]bool),
	}
}

func (f *Filter) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	f.mu.RLock()
	defer f.mu.RUnlock()

	rw := &ResponseWriter{ResponseWriter: w, Filter: f, state: state}
	return plugin.NextOrFailure(f.Name(), f.Next, ctx, rw, r)
}

func (f *Filter) Name() string {
	return "filter"
}
