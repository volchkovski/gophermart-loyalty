package httpserver

import "net/http"

type HTTPServer struct {
	router http.Handler
	addr   string
	notify chan error
}

func New(addr string, router http.Handler) *HTTPServer {
	return &HTTPServer{
		router: router,
		addr:   addr,
		notify: make(chan error, 1),
	}
}

func (s *HTTPServer) Start() {
	go func() {
		s.notify <- http.ListenAndServe(s.addr, s.router)
		close(s.notify)
	}()
}

func (s *HTTPServer) Notify() chan error {
	return s.notify
}
