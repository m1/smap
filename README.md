# smap

[![GoDoc](https://godoc.org/github.com/m1/smap?status.svg)](https://godoc.org/github.com/m1/smap)
[![Build Status](https://travis-ci.org/m1/smap.svg?branch=master)](https://travis-ci.org/m1/smap)
[![Go Report Card](https://goreportcard.com/badge/github.com/m1/smap)](https://goreportcard.com/report/github.com/m1/smap)
[![Release](https://img.shields.io/github/release/m1/smap.svg)](https://github.com/m1/smap/releases/latest)
[![Coverage Status](https://coveralls.io/repos/github/m1/smap/badge.svg)](https://coveralls.io/github/m1/smap)

## Installation

Use go get to get the latest version
```text
go get github.com/m1/smap
```

Then import it into your projects using the following:
```go
import (
	"github.com/m1/smap"
)
```

## Usage

smap can be used as a library, for example:

```go
c, _ := client.New(&client.Config{
    MaxWorkers:      50,
    IgnoreRobotsTxt: false,
    UserAgent:       "user-agent 1.1",
})
u, _ := url.Parse("http://example.com")
siteMap, err := c.Crawl(u)
for _, v := range siteMap {
	println(v.URL.path, len(v.Links), len(v.LinkedFrom))
}
```

## CLI usage
 
smap can also be used on the cli, just install using: `go get github.com/m1/gospin/cmd/gospin`

To use:
```
➜  ~ smap --help                    
smap is a site-mapping engine written in Go.

Usage:
  smap [url] [flags]

Flags:
  -h, --help                help for smap
      --json                json output
      --robots              Ignores robots.txt
  -u, --user-agent string   User agent to use for the crawler
  -v, --verbose             verbose printing
  -w, --workers int         How many workers to use (default 50)
```

For example:

```
➜  smap go build && ./smap http://google.com --json --verbose --workers=50 --user-agent="test-test" | jq
   {
     "/": {
       "path": "/",
       "redirects_to": null,
       "links": [
         "/advanced_search",
         "/language_tools",
         "/intl/en/ads/",
         "/services/",
         "/intl/en/policies/privacy/",
         "/intl/en/policies/terms/"
       ],
       "linked_from": [
         "/intl/en/ads/",
         "/services/",
         "/advanced_search",
         "/language_tools",
       ],
       "is_redirect": false
     }...
```