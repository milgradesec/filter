# Filter plugin

[![Build Status](https://travis-ci.org/milgradesec/filter.svg?branch=master)](https://travis-ci.org/milgradesec/filter)
[![codecov](https://codecov.io/gh/milgradesec/filter/branch/master/graph/badge.svg)](https://codecov.io/gh/milgradesec/filter)
[![Go Report Card](https://goreportcard.com/badge/github.com/milgradesec/filter)](https://goreportcard.com/report/github.com/milgradesec/filter)

CoreDNS plugin that blocks requests based on lists and rules

## Usage

~~~ corefile
.:53 {
    filter {
        allow ./lists/whitelist.txt
        block ./lists/blacklist.txt
    }
    forward . 1.1.1.1
}
~~~

## Building

~~~
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

~~~
$ go generate && go build
~~~
