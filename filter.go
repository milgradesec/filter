package filter

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"net/http"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/plugin/pkg/reuseport"
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
	ttl       uint32

	ln  net.Listener
	mux *http.ServeMux
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
	ln, err := reuseport.Listen("tcp", ":8080")
	if err != nil {
		return err
	}

	f.ln = ln
	f.mux = http.NewServeMux()

	f.mux.HandleFunc("/reload", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, "Reloading lists")
		err := f.Reload()
		if err != nil {
			fmt.Fprintf(w, "Reload failed: %v", err)
			return
		}
		fmt.Fprint(w, "Lists reloaded successfully")
	})

	go func() {
		http.Serve(f.ln, f.mux) //nolint
	}()

	return f.Load()
}

func (f *Filter) OnShutdown() error {
	return f.ln.Close()
}

func (f *Filter) Name() string {
	return "filter"
}
