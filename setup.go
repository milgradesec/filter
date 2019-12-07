package filter

import (
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
	filter := New()

	for c.Next() {
		for c.NextBlock() {
			switch c.Val() {
			case "allow":
				args := c.RemainingArgs()
				if len(args) != 1 {
					return nil, plugin.Error("filter", c.ArgErr())
				}
				filter.Lists[args[0]] = false

			case "block":
				args := c.RemainingArgs()
				if len(args) != 1 {
					return nil, plugin.Error("filter", c.ArgErr())
				}
				filter.Lists[args[0]] = true

			default:
				return nil, plugin.Error("filter", c.ArgErr())
			}
		}
	}
	if c.NextArg() {
		return nil, plugin.Error("filter", c.ArgErr())
	}

	if len(filter.Lists) == 0 {
		return nil, plugin.Error("filter", c.ArgErr())
	}
	return filter, nil
}
