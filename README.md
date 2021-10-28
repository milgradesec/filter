# filter

![CI](https://github.com/milgradesec/filter/workflows/CI/badge.svg)
[![CodeQL](https://github.com/milgradesec/filter/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/milgradesec/filter/actions/workflows/codeql-analysis.yml)
[![codecov](https://codecov.io/gh/milgradesec/filter/branch/main/graph/badge.svg)](https://codecov.io/gh/milgradesec/filter)
[![Go Report Card](https://goreportcard.com/badge/milgradesec/filter)](https://goreportcard.com/badge/github.com/milgradesec/filter)
[![Go Reference](https://pkg.go.dev/badge/github.com/milgradesec/filter.svg)](https://pkg.go.dev/github.com/milgradesec/filter)
![GitHub](https://img.shields.io/github/license/milgradesec/filter)

## Description

The _filter_ plugins enables blocking requests based on predefined lists and rules, creating a DNS sinkhole similar to Pi-Hole or AdGuard.

## Features

- Regex and simple string matching support.
- Inspection of CNAME, SVCB and HTTPS records detects and blocks cloaking.
- Block replies are fully cacheable by the _cache_ plugin.

## Syntax

```corefile
filter {
    allow FILE
    block FILE
    uncloak
    ttl DURATION
}
```

- `allow` load **FILE** to the whitelist.
- `block` load **FILE** to the blacklist.
- `uncloak` enables response uncloaking, disabled by default.
- `ttl` sets **TTL** for blocked responses, default is 3600s.

## Metrics

If monitoring is enabled (via the _prometheus_ plugin) then the following metric are exported:

- `coredns_filter_blocked_requests_total{server}` - count per server

## Examples

```corefile
.:53 {
    filter {
        allow /lists/allowlist.txt
        block /lists/denylist.txt
        uncloak
        ttl 600
    }
    forward . tls://1.1.1.1 tls://1.0.0.1 {
        tls_servername cloudflare-dns.com
    }
}
```
