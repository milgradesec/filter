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
- Load allow/block/allow-ips/block-ips from file or S3 bucket
- The allow-ips will only allow networks if a block is made with a smaller network prefix, like 0.0.0.0/0 or ::/0 as the default is allow.

## Environment

To use S3 buckets, the environment variables must be set: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, and `AWS_REGION`.  Once set the bucket path must start with `s3::`.


## Syntax

```corefile
filter {
    allow FILE
    block FILE
    allow-ips FILE
    block-ips FILE
    uncloak
    empty
    ttl DURATION
    reload DURATION
}
```

- `allow` load **FILE** to the whitelist.
- `block` load **FILE** to the blacklist.
- `allow-ips` load **FILE** to the IP response whitelist.
- `block-ips` load **FILE** to the IP response blacklist.
- `empty` return an empty answer record for every blocked request instead of an all zero record.
- `reload` **DURATION** read in the allow/block lists periodically, (example: reload 15s).
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
        block-ips /lists/bad-ips.txt
        uncloak
        ttl 600
    }
    forward . tls://1.1.1.1 tls://1.0.0.1 {
        tls_servername cloudflare-dns.com
    }
}
```
