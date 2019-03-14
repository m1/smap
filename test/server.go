package test

import (
	"fmt"
	"github.com/m1/smap/test/mock"
	"net/http"
	"net/http/httptest"
	"net/url"
)

type Server struct {
	Url *url.URL
	*http.ServeMux
	*httptest.Server
}

func NewServer() *Server {
	server := &Server{
		ServeMux: http.NewServeMux(),
	}
	server.HandleFunc("/", testResponse(mock.OkIndex))
	server.HandleFunc("/1/", testResponse(mock.OkPage1))
	server.HandleFunc("/2/", testResponse(mock.OkPage2))
	return server
}

func NewFailServer() *Server {
	server := &Server{
		ServeMux: http.NewServeMux(),
	}
	server.HandleFunc("/", testResponse(mock.FailIndex))
	server.HandleFunc("/1/", testResponse(mock.FailPage1))
	server.HandleFunc("/3/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		return
	})
	return server
}

func NewRedirectServer() *Server {
	server := &Server{
		ServeMux: http.NewServeMux(),
	}
	server.HandleFunc("/", testResponse(mock.OkIndex))
	server.HandleFunc("/1/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/2/", http.StatusMovedPermanently)
	})
	server.HandleFunc("/2/", testResponse(mock.OkPage2Redirected))
	return server
}

func NewBufErrServer() *Server {
	server := &Server{
		ServeMux: http.NewServeMux(),
	}
	server.HandleFunc("/2/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1")
	})
	return server
}

func testResponse(page string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, page)
		return
	}
}

func (s *Server) WithOkayRobots() {
	s.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprint(w, mock.RobotsDisallow)
		return
	})
}

func (s *Server) Start() *Server {
	s.Server = httptest.NewServer(s.ServeMux)

	u, err := url.Parse(s.URL)
	if err != nil {
		panic(err)
	}

	s.Url = u
	return s
}

func (s *Server) WithBufErrRobots() {
	s.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1")
	})
}

func (s *Server) With500ErrRobots() {
	s.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})
}
