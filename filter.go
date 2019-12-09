package filter

import (
	"context"
	"errors"
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
	allow *list
	block *list
	mu    sync.RWMutex

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

	f.mu.RLock()
	defer f.mu.RUnlock()

	rw := &ResponseWriter{ResponseWriter: w, Filter: f, state: state}
	return plugin.NextOrFailure(f.Name(), f.Next, ctx, rw, r)
}

func (f *Filter) Name() string {
	return "filter"
}
