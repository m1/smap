package crawler

import (
	"bytes"
	"fmt"
	"github.com/m1/smap/worker"
	"github.com/temoto/robotstxt"
	"io"
	"net/http"
	"net/url"
	"sync"
)

const (
	allowAllRobotsTxt     = "User-agent: *\nAllow: /"
	emailProtectionString = `/cdn-cgi/l/email-protection`
	contentTypeHTML       = "text/html"

	headerUserAgent   = "User-Agent"
	headerContentType = "Content-Type"

	tokenAnchor = "a"
	attrHref    = "href"
)

// Crawler is what handles the crawling of the pages including
// the result and urls to be consumed queues. Also handles
// and stores the worker pools and the final sitemap
type Crawler struct {
	// url is the url that is the base url to start the crawl at
	url url.URL

	// queued is the urls that are to be crawled, but
	// haven't been crawled yet
	queued sync.Map

	ignoreRobotsTxt bool
	robotsTxtParser *robotstxt.RobotsData
	userAgent       string

	// queue is the url chan to be consumed by the worker pool
	queue chan url.URL

	// results is the channel that handles the crawled pages
	results chan Page

	// SiteMap is the final result of the crawl
	SiteMap SiteMap

	// pagesWithErr are the pages that should be removed from
	// the sitemap
	pagesWithErr map[string]bool
	pool         *worker.Pool

	jobsCreated   int
	jobsCompleted int
}

// New passes back a new instance of a crawler
func New(u url.URL, ignoreRobotsTxt bool, maxWorkers int, userAgent string) *Crawler {
	u.Path = "/"
	return &Crawler{
		url:             u,
		ignoreRobotsTxt: ignoreRobotsTxt,
		queued:          sync.Map{},
		pool:            worker.NewPool(maxWorkers),
		pagesWithErr:    make(map[string]bool),
		SiteMap:         make(map[string]*Page),
		userAgent:       userAgent,
	}
}

// Run starts the crawling, and setting up/closing of the channels
func (c *Crawler) Run() error {
	if !c.ignoreRobotsTxt {
		err := c.robotsInit()
		if err != nil {
			return err
		}
	}
	c.queue = make(chan url.URL)
	defer close(c.queue)

	c.results = make(chan Page)
	defer close(c.results)

	c.pool.Start()
	c.addJob(c.url)

	c.waitForResults()
	c.cleanUpResults()

	c.generateLinksFrom()

	return nil
}

// robotsInit tries to fetch the robots.txt of the base url
func (c *Crawler) robotsInit() error {
	resp, err := http.Get(fmt.Sprintf("%s://%s/robots.txt", c.url.Scheme, c.url.Host))
	if err != nil {
		return err
	}

	body := resp.Body.(io.Reader)
	if resp.StatusCode > 400 {
		body = bytes.NewReader([]byte(allowAllRobotsTxt))
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(body)
	if err != nil {
		return err
	}

	c.robotsTxtParser, err = robotstxt.FromBytes(buf.Bytes())
	return err
}

func (c *Crawler) addJob(url url.URL) {
	page := NewPage(url, c)

	c.pool.AddJob(page)
	c.jobsCreated++
}

func (c *Crawler) waitForResults() {
	for {
		select {
		case page := <-c.queue:
			c.addJob(page)
		case result := <-c.results:
			if result.err != nil {
				c.pagesWithErr[result.URL.Path] = true
			} else {
				res := &result
				c.SiteMap[result.URL.Path] = res
			}

			c.jobsCompleted++
			if c.jobsCreated == c.jobsCompleted {
				c.pool.Close()
				return
			}
		}
	}
}

// cleanUpResults removes links from the SiteMap that
// produced errors once crawled
func (c *Crawler) cleanUpResults() {
	for path, page := range c.SiteMap {
		var errChecked links
		for _, link := range page.Links {
			exists, ok := c.pagesWithErr[link.URL.Path]
			if !ok || !exists {
				errChecked = append(errChecked, link)
			}
		}
		c.SiteMap[path].Links = errChecked
	}
}

// generateLinksFrom traverses through the crawled
// pages and finds where pages were linked from
// and stores them
func (c *Crawler) generateLinksFrom() {
	for _, page := range c.SiteMap {
		for _, l := range page.Links {
			c.SiteMap[l.URL.Path].appendLinkedFrom(link{
				URL: page.URL,
			})
		}
	}
}
