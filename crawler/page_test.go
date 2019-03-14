package crawler

import (
	"github.com/m1/smap/test"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func NewTestPage(parentUrl *url.URL, pageUrl *url.URL) (*Page, error) {
	crawler := New(*parentUrl, true, 1, "")
	return &Page{
		URL:        *pageUrl,
		crawler:    crawler,
		LinkedFrom: links{},
	}, nil
}

func TestPage_Crawl(t *testing.T) {
	s := test.NewServer()
	s.Start()
	defer s.Close()

	childUrl := s.Url
	childUrl.Path = "/2/"
	tc := testCase{
		Path:  childUrl.Path,
		Links: []string{"/", "/1/"},
	}

	tp, err := NewTestPage(s.Url, childUrl)
	if err != nil {
		t.Error(err)
	}
	tp.crawl()
	assert.Equal(t, tc.Path, tp.URL.Path)
	assert.Equal(t, tp.Links.Paths(), tc.Links)
}

func TestPage_Crawl_BufErr(t *testing.T) {
	s := test.NewBufErrServer()
	s.Start()
	defer s.Close()

	childUrl := s.Url
	childUrl.Path = "/2/"
	tp, err := NewTestPage(s.Url, childUrl)
	if err != nil {
		t.Error(err)
	}
	tp.crawl()
	if tp.err == nil {
		t.Error("should be error")
	}

	assert.EqualError(t, tp.err, "invalid response")
}

func TestPage_Crawl_NewReqErr(t *testing.T) {
	s := test.NewBufErrServer()
	s.Start()
	defer s.Close()

	childUrl := s.Url
	childUrl.Scheme = "---"
	childUrl.Path = "---"
	childUrl.Host = "---"
	tp, err := NewTestPage(s.Url, childUrl)
	if err != nil {
		t.Error(err)
	}
	tp.crawl()
	if tp.err == nil {
		t.Error("should be error")
	}

	assert.Contains(t, tp.err.Error(), "URL cannot contain colon")
}

func TestPage_Crawl_ClientDoErr(t *testing.T) {
	s := test.NewBufErrServer()
	s.Start()
	defer s.Close()

	childUrl := s.Url
	childUrl.Scheme = "ftp://"
	tp, err := NewTestPage(s.Url, childUrl)
	if err != nil {
		t.Error(err)
	}
	tp.crawl()
	if tp.err == nil {
		t.Error("should be error")
	}

	assert.Contains(t, tp.err.Error(), "unsupported protocol scheme")
}

func TestPage_parseLink_NotSameHost(t *testing.T) {
	s := test.NewBufErrServer()
	s.Start()
	defer s.Close()

	childUrl := s.Url
	childUrl.Host = "example.com"
	tp, err := NewTestPage(s.Url, childUrl)
	if err != nil {
		t.Error(err)
	}
	_, ok := tp.parseLink(childUrl, *s.Url, nil)
	if ok {
		t.Error("should error")
		return
	}
}

func TestPage_LinkValid_Err(t *testing.T) {
	s := test.NewBufErrServer()
	s.Start()
	defer s.Close()

	childUrl := s.Url
	childUrl.Path = emailProtectionString
	tp, err := NewTestPage(s.Url, childUrl)
	if err != nil {
		t.Error(err)
	}
	ok := tp.linkValid(childUrl, *s.Url)
	if ok {
		t.Error("should error")
		return
	}
}