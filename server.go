package webh

import (
	"context"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	*http.Server
}

// Start runs ListenAndServe on the http.Server with graceful shutdown.
func (s *Server) Start() {
	log.Info().Msgf("Server is running on port %s", s.Addr)
	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Msgf("closed Server error %s", err.Error())
		}
	}()
	s.GracefulShutdown()
}

func (s *Server) GracefulShutdown() {
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT)
	sig := <-quit
	log.Info().Msgf("Server is shutting down %s", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.SetKeepAlivesEnabled(false)
	if err := s.Shutdown(ctx); err != nil {
		log.Error().Msgf("could not gracefully shutdown the Server %s", err.Error())
	}
	log.Info().Msg("Server stopped")
}

type WebHandler func(w http.ResponseWriter, r *http.Request) error

func Unwrap(f WebHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)

		if err != nil {
			requestID := middleware.GetReqID(r.Context())
			log.Error().Caller(1).Err(err).Str("RequestID", requestID).Msg("cannot process request")
			ReturnErr(w, err)
		}
	}
}
