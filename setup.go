package filter

import (
	"strconv"

	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

func init() {
	plugin.Register("filter", setup)
}

func setup(c *caddy.Controller) error {
	p, err := parseConfig(c)
	if err != nil {
		return plugin.Error("filter", err)
	}

	c.OnStartup(func() error {
		//once.Do(func() { metrics.MustRegister(c, blockCount) })
		return p.Load()
	})

	c.OnShutdown(func() error {
		//close(block.stop)
		return nil
	})

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		p.Next = next
		return p
	})

	return nil
}

func parseConfig(c *caddy.Controller) (*Filter, error) {
	f := New()
	for c.Next() {
		for c.NextBlock() {
			if err := parseBlock(c, f); err != nil {
				return nil, err
			}
		}
	}

	if len(f.lists) == 0 {
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
		f.lists[c.Val()] = false

	case "block":
		if !c.NextArg() {
			return c.ArgErr()
		}
		f.lists[c.Val()] = true

	case "ttl":
		if !c.NextArg() {
			return c.ArgErr()
		}

		ttl, err := strconv.Atoi(c.Val())
		if err != nil {
			return err
		}
		f.ttl = uint32(ttl)

	default:
		return c.Errf("unknown setting '%s' ", c.Val())
	}
	return nil
}
