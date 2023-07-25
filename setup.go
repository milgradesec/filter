package filter

import (
	"strconv"
	"time"

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

func periodicFilterUpdate(f *Filter) chan bool {
	parseChan := make(chan bool)

	if f.reload == 0 {
		return parseChan
	}

	go func() {
		ticker := time.NewTicker(f.reload)
		defer ticker.Stop()
		for {
			select {
			case <-parseChan:
				return
			case <-ticker.C:
				if f.checkHash() {
					err := f.Load()
					if err != nil {
						log.Error(err)
					}
				}
			}
		}
	}()
	return parseChan
}

func setup(c *caddy.Controller) error {
	f, err := parseFilter(c)
	if err != nil {
		return plugin.Error(pluginName, err)
	}

	parseChan := periodicFilterUpdate(f)

	c.OnStartup(func() error {
		return f.Load()
	})

	c.OnShutdown(func() error {
		close(parseChan)
		return nil
	})

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

	return f, nil
}

func parseBlock(c *caddy.Controller, f *Filter) error {
	switch c.Val() {
	case "allow":
		if !c.NextArg() {
			return c.ArgErr()
		}

		l := listSource{Path: c.Val(), IsBlock: false, IsCIDR: false}
		f.sources = append(f.sources, l)

	case "block":
		if !c.NextArg() {
			return c.ArgErr()
		}

		l := listSource{Path: c.Val(), IsBlock: true, IsCIDR: false}
		f.sources = append(f.sources, l)

	case "allow-ips":
		if !c.NextArg() {
			return c.ArgErr()
		}

		l := listSource{Path: c.Val(), IsBlock: false, IsCIDR: true}
		f.sources = append(f.sources, l)

	case "block-ips":
		if !c.NextArg() {
			return c.ArgErr()
		}

		l := listSource{Path: c.Val(), IsBlock: true, IsCIDR: true}
		f.sources = append(f.sources, l)

	case "reload":
		if !c.NextArg() {
			return c.ArgErr()
		}
		d, err := time.ParseDuration(c.Val())
		if err != nil {
			return c.Errf("invalid reload value: %s", c.Val())
		}
		f.reload = d

	case "empty":
		if c.NextArg() {
			return c.ArgErr()
		}
		f.emptyResponse = true

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
