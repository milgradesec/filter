package filter

import (
	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
)

const pluginName = "filter"

func init() {
	plugin.Register(pluginName, setup)
}

func setup(c *caddy.Controller) error {
	f, err := parseFilter(c)
	if err != nil {
		return plugin.Error(pluginName, err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		f.Next = next
		return f
	})

	c.OnStartup(func() error {
		metrics.MustRegister(c, BlockCount)
		return f.Load()
	})

	return nil
}

func parseFilter(c *caddy.Controller) (*Filter, error) {
	f := New()

	for c.Next() {
		for c.NextBlock() {
			if err := parseBlock(c, f); err != nil {
				return nil, err
			}
		}
	}

	return f, nil
}

func parseBlock(c *caddy.Controller, f *Filter) error {
	switch c.Val() {
	case "allow":
		if !c.NextArg() {
			return c.ArgErr()
		}

		l := source{Path: c.Val(), Block: false}
		f.sources = append(f.sources, l)

	case "block":
		if !c.NextArg() {
			return c.ArgErr()
		}

		l := source{Path: c.Val(), Block: true}
		f.sources = append(f.sources, l)

	case "exclude":
		if c.NextArg() {
			return c.ArgErr()
		}

	case "uncloak_cname":
		if c.NextArg() {
			return c.ArgErr()
		}
		f.uncloak = true

	default:
		return c.Errf("unknown option '%s' ", c.Val())
	}
	return nil
}
