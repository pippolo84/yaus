package yaus

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/pippolo84/yaus/internal/short"
	"github.com/pippolo84/yaus/internal/storage"
	"go.uber.org/zap"
)

const (
	// DefaultWriteTimeout is the default server write timeout
	DefaultWriteTimeout = 15 * time.Second
	// DefaultReadTimeout is the default server read timeout
	DefaultReadTimeout = 15 * time.Second
	// DefaultIdleTimeout is the default server idle timeout
	DefaultIdleTimeout = 60 * time.Second
)

// Server is a HTTP server that supports graceful shutdown.
type Server struct {
	server    http.Server
	shortener short.Hasher
	cache     storage.Backend
	logger    *zap.SugaredLogger
}

// Option is a function able to set a configuration option on the Server
type Option func(*Server)

// Address returns an Option to set the specified addr as the listening
// address on the Server
func Address(addr string) Option {
	return func(s *Server) {
		s.server.Addr = addr
	}
}

// WriteTimeout returns an Option to set the specified write timeout on the Server
func WriteTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.server.WriteTimeout = d
	}
}

// ReadTimeout returns an Option to set the specified read timeout on the Server
func ReadTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.server.ReadTimeout = d
	}
}

// IdleTimeout returns an Option to set the specified idle timeout on the Server
func IdleTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.server.IdleTimeout = d
	}
}

// NewServer returns a reference to a new Server listening on addr.
func NewServer(
	cache storage.Backend,
	shortener short.Hasher,
	logger *zap.SugaredLogger,
	options ...Option,
) *Server {
	router := mux.NewRouter()

	// set default values
	srv := &Server{
		server: http.Server{
			WriteTimeout: DefaultWriteTimeout,
			ReadTimeout:  DefaultReadTimeout,
			IdleTimeout:  DefaultIdleTimeout,
			Handler:      router,
		},
		shortener: shortener,
		cache:     cache,
		logger:    logger,
	}

	// configure with given options
	for _, opt := range options {
		opt(srv)
	}

	router.HandleFunc("/{hash}", srv.redirect()).Methods(http.MethodGet)
	router.HandleFunc("/shorten", srv.shorten()).Methods(http.MethodPost)

	return srv
}

// Run starts the server, making it listening on specified address-
// It returns a channel where all errors are notified.
func (s *Server) Run() <-chan error {
	errs := make(chan error)

	go func() {
		defer close(errs)

		s.logger.Infow("server start listening", zap.String("address", s.server.Addr))

		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs <- err
		}

		s.logger.Info("server is stopping...")
	}()

	return errs
}

// Shutdown makes the server stop listening and refuse further connections.
// It takes a context to limit the shutdown duration and a wait group to signal
// the caller when the shutdown process has finished.
func (s *Server) Shutdown(ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()
	s.server.SetKeepAlivesEnabled(false)
	return s.server.Shutdown(ctx)
}

func (s *Server) redirect() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)

		url, err := s.cache.Get(params["hash"])
		if err != nil {
			s.logger.Error("cache get", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func (s *Server) shorten() http.HandlerFunc {
	type reqBody struct {
		URL string `json:"url"`
	}
	type response struct {
		Hash string `json:"hash"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var rb reqBody
		if err := json.NewDecoder(r.Body).Decode(&rb); err != nil {
			s.logger.Error("shorten decode", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		key := s.shortener.Hash(rb.URL)

		if err := s.cache.Put(key, rb.URL); err != nil {
			s.logger.Error("cache put", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(response{key}); err != nil {
			s.logger.Error("shorten encode", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}
