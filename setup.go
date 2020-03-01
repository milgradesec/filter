package filter

import (
	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
)

func init() { plugin.Register("filter", setup) }

func setup(c *caddy.Controller) error {
	f, err := parseConfig(c)
	if err != nil {
		return plugin.Error("filter", err)
	}

	c.OnStartup(func() error {
		metrics.MustRegister(c, BlockCount)
		return f.OnStartup()
	})

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		f.Next = next
		return f
	})

	return nil
}

func parseConfig(c *caddy.Controller) (*Filter, error) {
	f := &Filter{}
	for c.Next() {
		for c.NextBlock() {
			if err := parseBlock(c, f); err != nil {
				return nil, err
			}
		}
	}

	if len(f.Lists) == 0 {
		return nil, c.ArgErr()
	}
	return f, nil
}

func parseBlock(c *caddy.Controller, f *Filter) error {
	switch c.Val() {
	case "allow":
		if !c.NextArg() {
			return c.ArgErr()
		}
		l := &List{Path: c.Val(), Block: false}
		f.Lists = append(f.Lists, l)

	case "block":
		if !c.NextArg() {
			return c.ArgErr()
		}
		l := &List{Path: c.Val(), Block: true}
		f.Lists = append(f.Lists, l)

	case "uncloak":
		f.uncloak = true

	default:
		return c.Errf("unknown setting '%s' ", c.Val())
	}
	return nil
}
