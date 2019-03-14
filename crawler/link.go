package crawler

import (
	"net/url"
)

// link small struct to store the crawling url/link
// and a bool to tell if the link has been crawled
type link struct {
	URL     url.URL
	Crawled bool
}

type links []link

func (l links) Paths() []string {
	var links []string
	for _, link := range l{
		links = append(links, link.URL.Path)
	}
	return links
}
