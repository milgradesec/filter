package filter

import (
	"github.com/coredns/coredns/plugin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	BlockedCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: "filter",
		Name:      "blocked_total",
		Help:      "The count of blocked requests.",
	}, []string{"server"})
)
