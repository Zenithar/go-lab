package main

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/soheilhy/cmux"
	"go.uber.org/zap"
)

type exampleHTTPHandler struct{}

func (h *exampleHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "example http response")
}

func serveHTTP(l net.Listener) {
	s := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      &exampleHTTPHandler{},
	}
	if err := s.Serve(l); err != cmux.ErrListenerClosed {
		logger.Panic("Unable to start HTTP Server", zap.Error(err))
	}
}
