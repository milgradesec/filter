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

type Filter struct {
	Next plugin.Handler

	Lists          []*List //map[string]bool
	BlockedTtl     uint32
	ReloadInterval time.Duration

	whitelist *PatternMatcher
	blacklist *PatternMatcher

	//ln   net.Listener
	//mux  *http.ServeMux
	stop chan bool
}

func (f *Filter) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	name := trimTrailingDot(state.Name())

	if f.Match(name) {
		BlockCount.WithLabelValues(metrics.WithServer(ctx)).Inc()
		return writeNXdomain(w, r)
	}
	return plugin.NextOrFailure(f.Name(), f.Next, ctx, w, r)
}

func (f *Filter) Match(str string) bool {
	if f.whitelist.Match(str) {
		return false
	}
	if f.blacklist.Match(str) {
		return true
	}
	return false
}

func (f *Filter) OnStartup() error {
	/*ln, err := reuseport.Listen("tcp", "8080")
	if err != nil {
		return err
	}
	f.ln = ln
	f.mux = http.NewServeMux()

	f.mux.HandleFunc("/reload", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "Reloading...") //nolint
	})

	go func() {
		http.Serve(f.ln, f.mux) //nolint
	}()*/
	f.stop = make(chan bool)
	return f.Load()
}

func (f *Filter) OnShutdown() error {
	close(f.stop)
	return nil
}

func (f *Filter) Name() string {
	return "filter"
}
