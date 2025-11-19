package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Estriper0/avito_intership/internal/config"
)

type Server struct {
	httpServer *http.Server
	config     *config.Config
	err        chan error
}

func New(handler http.Handler, config *config.Config) *Server {
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Server.Port),
		Handler:      handler,
		ReadTimeout:  config.Server.ReadTimeout,
		WriteTimeout: config.Server.WriteTimeout,
	}

	return &Server{
		httpServer: server,
		config:     config,
		err:        make(chan error, 1),
	}
}

func (s *Server) Err() <-chan error {
	return s.err
}

func (s *Server) Run() {
	s.err <- s.httpServer.ListenAndServe()
	close(s.err)
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Server.ShutdownTimeout)
	defer cancel()

	return s.httpServer.Shutdown(ctx)
}
