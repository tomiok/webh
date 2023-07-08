package webh

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	s *http.Server
	*chi.Mux
}

// NewServer return a *Server.
func NewServer(port, serviceName string) *Server {
	srv := newServer(port)
	logger := httplog.NewLogger(serviceName, httplog.Options{
		JSON:            true,
		Concise:         true,
		TimeFieldFormat: time.UnixDate,
	})

	srv.Use(
		httplog.RequestLogger(logger),
		middleware.Heartbeat("/ping"),
	)
	return srv
}

func newServer(port string) *Server {
	mux := chi.NewMux()
	return &Server{
		Mux: mux,
		s: &http.Server{
			Addr:         ":" + port,
			WriteTimeout: time.Second * 15,
			ReadTimeout:  time.Second * 15,
			IdleTimeout:  time.Second * 60,
			Handler:      mux,
		},
	}
}

// Start runs ListenAndServe on the http.Server with graceful shutdown.
func (s *Server) Start() {
	log.Info().Msgf("server is running on port %s", s.s.Addr)
	go func() {
		if err := s.s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Msgf("closed Server error %s", err.Error())
		}
	}()
	s.gracefulShutdown()
}

func (s *Server) gracefulShutdown() {
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT)
	sig := <-quit
	log.Info().Msgf("server is shutting down %s", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.s.SetKeepAlivesEnabled(false)
	if err := s.s.Shutdown(ctx); err != nil {
		log.Error().Msgf("could not gracefully shutdown the Server %s", err.Error())
	}
	log.Info().Msg("server stopped")
}

type WebHandler func(w http.ResponseWriter, r *http.Request) error

func Unwrap(h WebHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)

		if err != nil {
			oplog := httplog.LogEntry(r.Context())

			requestID := middleware.GetReqID(r.Context())
			evt := oplog.Error()
			if requestID != "" {
				evt = evt.Str("RequestID", requestID)
			}

			evt.Msg("cannot process request")
			wrapErrorResponse(w, requestID, err)
		}
	}
}
