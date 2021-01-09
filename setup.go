package filter

import (
	"strconv"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
)

const pluginName = "filter"

var log = clog.NewWithPlugin(pluginName)

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

	err := f.Load()
	if err != nil {
		return nil, err
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

	case "uncloak":
		if c.NextArg() {
			return c.ArgErr()
		}
		f.uncloak = true

	case "ttl":
		if !c.NextArg() {
			return c.ArgErr()
		}

		ttl, err := strconv.ParseUint(c.Val(), 10, 32)
		if err != nil {
			return c.Errf("invalid ttl value: %s", c.Val())
		}
		f.ttl = uint32(ttl)

	default:
		return c.Errf("unknown option '%s' ", c.Val())
	}
	return nil
}
