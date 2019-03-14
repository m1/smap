package crawler

import (
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Page is the type that stores all the links to and rom a
// url/page. Also stores if there were errors when crawling the page
type Page struct {
	URL         url.URL  `json:"-"`
	Links       links    `json:"links"`
	LinkedFrom  links    `json:"linked_from"`
	IsRedirect  bool     `json:"is_redirect"`
	RedirectsTo *url.URL `json:"redirects_to"`

	crawler *Crawler `json:"-"`
	err     error    `json:"-"`
}

// MarshalJSON generates the json and converts the `link`
// and `LinkedFrom` from type `links` to `[]string` as well
// as passing back the `Path` and `RedirectsTo` as `string`
func (p *Page) MarshalJSON() ([]byte, error) {
	var links []string
	var linkedFrom []string

	for _, l := range p.Links {
		links = append(links, l.URL.Path)
	}

	for _, l := range p.LinkedFrom {
		linkedFrom = append(linkedFrom, l.URL.Path)
	}

	var redirectsTo *string
	if p.IsRedirect {
		redirectPath := p.RedirectsTo.Path
		redirectsTo = &redirectPath
	}

	type Alias Page
	return json.Marshal(&struct {
		Path        string   `json:"path"`
		RedirectsTo *string  `json:"redirects_to"`
		Links       []string `json:"links"`
		LinkedFrom  []string `json:"linked_from"`
		*Alias
	}{
		Path:        p.URL.Path,
		RedirectsTo: redirectsTo,
		Links:       links,
		LinkedFrom:  linkedFrom,
		Alias:       (*Alias)(p),
	})
}

// NewPage passes back a new instance of `Page`
func NewPage(url url.URL, crawler *Crawler) *Page {
	return &Page{
		URL:        url,
		crawler:    crawler,
		LinkedFrom: links{},
	}
}

// Run is what gets called when a worker starts work/crawling
// on a page
func (p Page) Run() {
	p.crawl()
	if p.err == nil {
		for _, link := range p.Links {
			if !link.Crawled {
				p.crawler.queue <- link.URL
			}
		}
	}

	p.crawler.results <- p
}

// crawl is the main body of work for `Page`, it fetches
// the body of the url and tries parsing the html and scanning
// for links within the page
func (p *Page) crawl() {
	linkCache := make(map[string]bool)

	resp, err := p.makeRequest(p.URL)
	if err != nil {
		p.err = err
		return
	}
	defer resp.Body.Close()
	tokenizer := html.NewTokenizer(resp.Body)

	if p.URL.Path != resp.Request.URL.Path ||
		p.URL.Host != resp.Request.URL.Host {
		p.IsRedirect = true
		p.RedirectsTo = resp.Request.URL
	}
	for {
		t := tokenizer.Next()
		if t == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				break
			}

			p.err = errors.New("error parsing html")
			return
		}
		token := tokenizer.Token()
		if token.DataAtom.String() == tokenAnchor {
			for _, attr := range token.Attr {
				if attr.Key == attrHref {
					linkURL, err := p.getLinkURL(attr)
					if err != nil {
						continue
					}
					link, ok := p.parseLink(linkURL, p.URL, &linkCache)
					if !ok {
						continue
					}
					p.Links = append(p.Links, link)
				}
			}
		}
	}

	return
}

// makeRequest sets up the http client and does the
// GETing of the page url, also checks for a successful
// response
func (p *Page) makeRequest(u url.URL) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set(headerUserAgent, p.crawler.userAgent)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	ct := resp.Header.Get(headerContentType)
	if resp.StatusCode >= http.StatusOK &&
		resp.StatusCode < http.StatusBadRequest &&
		strings.Contains(ct, contentTypeHTML) {
		return resp, nil
	}

	return nil, errors.New("invalid response")
}

// getLinkURL parses and gets a new link from the href
// on the anchor attribute
func (p *Page) getLinkURL(attr html.Attribute) (*url.URL, error) {
	v := strings.TrimLeft(attr.Val, "/")
	if !strings.HasSuffix(v, "/") {
		v = v + "/"
	}

	linkURL, err := url.Parse(v)
	if err != nil || !linkURL.IsAbs() {
		if v == "/" {
			v = ""
		}
		absLink := fmt.Sprintf("%s://%s/%s", p.crawler.url.Scheme, p.crawler.url.Host, v)
		linkURL, err = url.Parse(absLink)
		if err != nil {
			return nil, err
		}
	}

	return linkURL, err
}

// parseLink checks to see if the link has already been queued
// and validates the url to see if valid, checking for if
// we can crawl it in context of the robots.txt and also
// if the link is on the same host
func (p *Page) parseLink(linkURL *url.URL, parent url.URL, cache *map[string]bool) (link, bool) {
	link := link{
		URL:     *linkURL,
		Crawled: true,
	}

	if linkURL.Host != p.crawler.url.Host {
		return link, false
	}

	if !p.linkValid(linkURL, parent) {
		return link, false
	}

	if !p.crawler.ignoreRobotsTxt &&
		!p.crawler.robotsTxtParser.TestAgent(linkURL.Path, p.crawler.userAgent) {
		return link, false
	}

	path := linkURL.Path
	queued, ok := (*cache)[path]
	if !ok || !queued {
		queued, ok := p.crawler.queued.Load(path)
		if !ok || !queued.(bool) {
			link.Crawled = false
			p.crawler.queued.Store(path, true)
		}

		(*cache)[path] = true
	} else if queued {
		return link, false
	}

	return link, true
}

func (p *Page) linkValid(linkURL *url.URL, parent url.URL) bool {
	return !strings.Contains(linkURL.Path, emailProtectionString) &&
		linkURL.Path != parent.Path &&
		linkURL.String() != parent.String()
}

func (p *Page) appendLinkedFrom(link link) {
	p.LinkedFrom = append(p.LinkedFrom, link)
}
