package crawler

import (
	"github.com/m1/smap/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testCase struct {
	Path       string
	Links      []string
	LinkedFrom []string
	Redirect   bool
	RedirectTo string
}

func TestCrawler_Run(t *testing.T) {
	tc := make(map[string]testCase)
	tc["/"] = testCase{
		Path:       "/",
		Links:      []string{"/1/"},
		LinkedFrom: []string{"/2/"},
	}
	tc["/1/"] = testCase{
		Path:       "/1/",
		Links:      []string{"/2/"},
		LinkedFrom: []string{"/", "/2/"},
	}
	tc["/2/"] = testCase{
		Path:       "/2/",
		Links:      []string{"/", "/1/"},
		LinkedFrom: []string{"/1/"},
	}

	s := test.NewServer()
	s.Start()
	defer s.Close()

	crawler := New(*s.Url, true, 1, "")
	err := crawler.Run()
	if err != nil {
		t.Error(err)
	}

	for k, v := range tc {
		sm, ok := crawler.SiteMap[k]
		if !ok {
			t.Error("should exist")
		}

		assert.ElementsMatch(t, v.Links, sm.Links.Paths())
		assert.ElementsMatch(t, v.LinkedFrom, sm.LinkedFrom.Paths())
	}
}

func TestCrawler_Run_WithRedirect(t *testing.T) {
	tc := make(map[string]testCase)
	tc["/"] = testCase{
		Path:       "/",
		Links:      []string{"/1/"},
		LinkedFrom: []string{"/1/"},
		Redirect:   false,
	}
	tc["/1/"] = testCase{
		Path:       "/1/",
		Links:      []string{"/"},
		LinkedFrom: []string{"/"},
		Redirect:   true,
		RedirectTo: "/2/",
	}

	s := test.NewRedirectServer()
	s.Start()
	defer s.Close()

	crawler := New(*s.Url, true, 1, "")
	err := crawler.Run()
	if err != nil {
		t.Error(err)
	}

	for k, v := range tc {
		sm, ok := crawler.SiteMap[k]
		if !ok {
			t.Error("should exist")
		}

		assert.ElementsMatch(t, v.Links, sm.Links.Paths())
		assert.ElementsMatch(t, v.LinkedFrom, sm.LinkedFrom.Paths())
		assert.Equal(t, v.Redirect, sm.IsRedirect)

		if v.Redirect {
			assert.Equal(t, v.RedirectTo, sm.RedirectsTo.Path)
		}
	}
}

func TestCrawler_Run_WithLinksErr(t *testing.T) {
	tc := make(map[string]testCase)
	tc["/"] = testCase{
		Path:       "/",
		Links:      []string{"/1/"},
		LinkedFrom: []string{},
	}
	tc["/1/"] = testCase{
		Path:       "/1/",
		Links:      []string{},
		LinkedFrom: []string{"/"},
	}

	s := test.NewFailServer()
	s.Start()
	defer s.Close()

	crawler := New(*s.Url, true, 1, "")
	err := crawler.Run()
	if err != nil {
		t.Error(err)
	}

	for k, v := range tc {
		sm, ok := crawler.SiteMap[k]
		if !ok {
			t.Error("should exist")
		}

		assert.ElementsMatch(t, v.Links, sm.Links.Paths())
		assert.ElementsMatch(t, v.LinkedFrom, sm.LinkedFrom.Paths())
	}
}

func TestCrawler_RunWithRobotsTxt_Disallow(t *testing.T) {
	tc := make(map[string]testCase)
	tc["/"] = testCase{
		Path:       "/",
		Links:      []string{"/1/"},
		LinkedFrom: []string{},
	}
	tc["/1/"] = testCase{
		Path:       "/1/",
		Links:      []string{},
		LinkedFrom: []string{"/"},
	}

	s := test.NewServer()
	s.WithOkayRobots()
	s.Start()
	defer s.Close()

	crawler := New(*s.Url, false, 1, "test-robot")
	err := crawler.Run()
	if err != nil {
		t.Error(err)
	}

	for k, v := range tc {
		sm, ok := crawler.SiteMap[k]
		if !ok {
			t.Error("should exist")
		}

		assert.ElementsMatch(t, v.Links, sm.Links.Paths())
		assert.ElementsMatch(t, v.LinkedFrom, sm.LinkedFrom.Paths())
	}
}

func TestCrawler_RunWithRobotsTxt_BufError(t *testing.T) {
	s := test.NewServer()
	s.WithBufErrRobots()
	s.Start()
	defer s.Close()

	crawler := New(*s.Url, false, 1, "test-robot")
	err := crawler.Run()
	if err == nil {
		t.Error("expecting error")
	}

	assert.EqualError(t, err, "unexpected EOF")
}

func TestCrawler_RunWithRobotsTxt_500Error(t *testing.T) {
	s := test.NewServer()
	s.With500ErrRobots()
	s.Start()
	defer s.Close()

	crawler := New(*s.Url, false, 1, "test-robot")
	err := crawler.Run()
	if err != nil {
		t.Error(err)
	}

	assert.Empty(t, crawler.robotsTxtParser.Sitemaps)
}

func TestCrawler_RunWithRobotsTxt_GetError(t *testing.T) {
	s := test.NewServer()
	s.With500ErrRobots()
	s.Start()
	s.Url.Scheme = "ftp://"
	defer s.Close()

	crawler := New(*s.Url, false, 1, "test-robot")
	err := crawler.Run()
	if err == nil {
		t.Error(err)
	}

	assert.Contains(t, err.Error(), "unsupported protocol scheme")
}
