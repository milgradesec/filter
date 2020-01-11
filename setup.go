package filter

import (
	"strconv"

	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
)

func init() {
	plugin.Register("filter", setup)
}

func setup(c *caddy.Controller) error {
	f, err := parseConfig(c)
	if err != nil {
		return plugin.Error("filter", err)
	}

	c.OnStartup(func() error {
		metrics.MustRegister(c, BlockCount)
		return f.OnStartup()
	})

	c.OnShutdown(func() error {
		return f.OnShutdown()
	})

	c.OnRestart(func() error {
		return f.OnShutdown()
	})

	c.OnRestartFailed(func() error {
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

	case "ttl":
		if !c.NextArg() {
			return c.ArgErr()
		}

		ttl, err := strconv.Atoi(c.Val())
		if err != nil {
			return err
		}
		f.BlockedTtl = uint32(ttl)

	default:
		return c.Errf("unknown setting '%s' ", c.Val())
	}
	return nil
}
