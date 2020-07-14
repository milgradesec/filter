# filter

![CI](https://github.com/milgradesec/filter/workflows/CI/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/milgradesec/filter)](https://goreportcard.com/badge/github.com/milgradesec/filter)
[![Code Coverage](https://img.shields.io/codecov/c/github/milgradesec/filter/master.svg)](https://codecov.io/github/milgradesec/filter?branch=master)
![GitHub](https://img.shields.io/github/license/milgradesec/filter)

## Description

The _filter_ plugins enables blocking requests based on predefined lists and rules, creating a DNS sinkhole similar to Pi-Hole.

## Features

- Regex and simple string matching support.
- CNAME inspection detects and blocks cloaking.
- Block replies are fully cacheable by the _cache_ plugin.

## Syntax

```corefile
filter {
    allow FILE
    block FILE
    uncloak
}
```

- `allow` load **FILE** to the whitelist.
- `block` load **FILE** to the blacklist.
- `uncloak` enables cname uncloaking.

## Metrics

If monitoring is enabled (via the _prometheus_ plugin) then the following metric are exported:

- `coredns_filter_blocked_requests_total{server}` - count per server

## Examples

```corefile
.:53 {
    filter {
        allow ./lists/whitelist.txt
        block ./lists/blacklist.txt
        uncloak
    }
    forward . 1.1.1.1
}
```
