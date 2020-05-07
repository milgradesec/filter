package filter

import (
	"github.com/coredns/coredns/plugin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	BlockCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: pluginName,
		Name:      "blocked_requests_total",
		Help:      "Counter of blocked requests.",
	}, []string{"server"})
)
