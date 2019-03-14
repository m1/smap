package client

import (
	"github.com/go-errors/errors"
	"github.com/m1/smap/crawler"
	"net/url"
)

const (
	defaultUserAgent  = "smap-v0.0.1"
	defaultMaxWorkers = 50
)

// Client is the main entry point for this package, it
// enables you to define a crawling client then to crawl
// multiple urls with the same client/config
type Client struct {
	Config *Config
}

// Config is the config for the client
type Config struct {
	// MaxWorkers is the maximum numbers of workers for
	// the crawling pool to have, ideally will
	// be around 50-100
	MaxWorkers int

	// IgnoreRobotsTxt enables polite crawling to be turned
	// off, don't recommended setting this to true
	IgnoreRobotsTxt bool

	// UserAgent is the user agent that the crawler will use
	// defaults to `smap-v0.0.1`
	UserAgent string
}

// New passes back a new client, populates the config
// with the default values if not set
func New(config *Config) (*Client, error) {
	if config.MaxWorkers == 0 {
		config.MaxWorkers = defaultMaxWorkers
	}

	if config.MaxWorkers < 0 {
		return nil, errors.New("maxworkers must be above 0")
	}

	if config.UserAgent == "" {
		config.UserAgent = defaultUserAgent
	}

	return &Client{
		Config: config,
	}, nil
}

// Crawl starts the crawling of the url and passes back the
// sitemap
func (c *Client) Crawl(url *url.URL) (crawler.SiteMap, error) {
	if url.Path != "/" && url.Path != "" {
		return nil, errors.New("url should be the base url")
	}

	cr := crawler.New(*url, c.Config.IgnoreRobotsTxt, c.Config.MaxWorkers, c.Config.UserAgent)
	err := cr.Run()
	if err != nil {
		return nil, err
	}

	return cr.SiteMap, nil
}
