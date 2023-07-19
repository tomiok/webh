package webh

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type option func(options *options)
type Server struct {
	s *http.Server

	*chi.Mux
}

// options is for whatever server you want to build. The usability is production-ready but you can still
// add more middlewares or configurations.
type options struct {
	cors      func(http.Handler) http.Handler
	logger    func(http.Handler) http.Handler
	heartbeat func(http.Handler) http.Handler
}

func (o *options) use() []func(http.Handler) http.Handler {
	var mids []func(http.Handler) http.Handler

	if o.heartbeat != nil {
		mids = append(mids, o.heartbeat)
	}
	if o.logger != nil {
		mids = append(mids, o.logger)
	}
	if o.cors != nil {
		mids = append(mids, o.cors)
	}

	return mids
}

type CorsOpt struct {
	AllowedOrigins []string
	AllowedHeaders []string
	ExposedHeaders []string

	AllowCredentials bool
	MaxAge           int
}

// WithCors add cors to your web server.
func WithCors(opt CorsOpt) option {
	return func(options *options) {
		options.cors = cors.Handler(cors.Options{

			AllowedOrigins:   opt.AllowedOrigins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   opt.AllowedHeaders,
			ExposedHeaders:   opt.ExposedHeaders,
			AllowCredentials: opt.AllowCredentials,
			MaxAge:           opt.MaxAge, // Maximum value not ignored by any of major browsers
		})
	}
}

// WithLogger add a JSON logger to your web server.
func WithLogger(serviceName string) option {
	return func(opt *options) {
		logger := httplog.NewLogger(serviceName, httplog.Options{
			LogLevel:      "INFO",
			JSON:          true,
			Concise:       true,
			TimeFieldName: "at",
		})
		opt.logger = httplog.RequestLogger(logger)
	}
}

// WithHeartbeat add a dummy endpoint in the specified path. Will return a non-json response "." with a 200 status.
func WithHeartbeat(path string) option {
	return func(opt *options) {
		opt.heartbeat = middleware.Heartbeat(path)
	}
}

// NewServer return a *Server.
func NewServer(port string, serverOptions ...option) *Server {
	srv := newServer(port)
	opts := &options{}

	for _, opt := range serverOptions {
		opt(opts)
	}

	mids := opts.use()
	srv.Use(mids...)

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

			Handler: mux,
		},
	}
}

// Start runs ListenAndServe on the http.Server with graceful shutdown.
func (s *Server) Start() {
	log.Printf("server is running on port %s \n", s.s.Addr)
	go func() {
		if err := s.s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("closed Server error %s", err.Error())
		}
	}()
	s.gracefulShutdown()
}

func (s *Server) gracefulShutdown() {
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT)
	sig := <-quit
	log.Printf("server is shutting down %s", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.s.SetKeepAlivesEnabled(false)
	if err := s.s.Shutdown(ctx); err != nil {
		log.Printf("could not gracefully shutdown the Server %s", err.Error())
	}
	log.Printf("server stopped")
}

type WebHandler func(w http.ResponseWriter, r *http.Request) error

func Unwrap(h WebHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)

		if err != nil {
			oplog := httplog.LogEntry(r.Context())

			requestID := middleware.GetReqID(r.Context())
			oplog.Error().Str("RequestID", requestID).Msg("cannot process request")

			wrapErrorResponse(w, requestID, err)
		}
	}
}
