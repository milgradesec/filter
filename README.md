# filter

## Name

*filter* - enables blocking requests based on lists and rules.

## Description

## Syntax

## Features

* Regex and simple string matching
* Detects CNAME cloacking
* Responses allow negative caching

## Metrics

If monitoring is enabled (via the *prometheus* plugin) then the following metric are exported:

* `coredns_filter_blocked_requests_total{server}` - count per server

## Examples

~~~ corefile
.:53 {
    filter {
        allow ./lists/whitelist.txt
        block ./lists/blacklist.txt
        uncloak
    }
    forward . 1.1.1.1
}
~~~

## Building

~~~ cmd
$ git clone https://github.com/coredns/coredns
$ cd coredns
~~~

Then modify plugin.cfg.

~~~ txt
...
cache:cache
filter:github.com/milgradesec/filter
forward:forward
...
~~~

And build coredns as usual.

~~~ cmd
$ go generate && go build
~~~
