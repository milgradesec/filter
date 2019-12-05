package filter

import (
	"errors"
	"os"

	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	plugin.Register("filter", setup)
}

func setup(c *caddy.Controller) error {
	p, err := parseFilter(c)
	if err != nil {
		return plugin.Error("filter", err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		p.Next = next
		return p
	})

	c.OnStartup(func() error {
		m := dnsserver.GetConfig(c).Handler("prometheus")
		if x, ok := m.(*metrics.Metrics); ok {
			x.MustRegister(blockedCount)
		}
		return nil
	})
	return nil
}

func parseFilter(c *caddy.Controller) (*Plugin, error) {
	f := &Plugin{}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	for c.Next() {
		for c.NextBlock() {
			switch c.Val() {
			case "list":
				args := c.RemainingArgs()
				if len(args) != 2 {
					return nil, c.ArgErr()
				}

				list, err := NewFilter(cwd + args[0])
				if err != nil {
					return nil, err
				}

				switch args[1] {
				case "white":
					list.Type = white
				case "black":
					list.Type = black
				case "private":
					list.Type = private
				default:
					return nil, c.ArgErr()
				}

				f.filters = append(f.filters, list)
			default:
				return nil, c.ArgErr()
			}
		}
	}

	if len(f.filters) == 0 {
		return nil, errors.New("invalid list configuration")
	}
	return f, nil
}

var (
	blockedCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: "filter",
		Name:      "blocked_total",
		Help:      "The count of blocked requests.",
	}, []string{"server"})
)
