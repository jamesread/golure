// Package httpserve builds and runs net/http servers with consistent timeouts
// and graceful shutdown.
package httpserve

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	defaultReadHeaderTimeout = 10 * time.Second
	defaultIdleTimeout       = 120 * time.Second
	defaultShutdownTimeout   = 10 * time.Second
)

// Config describes a server. Zero timeouts use the defaults for
// ReadHeaderTimeout (10s), IdleTimeout (120s) and ShutdownTimeout (10s).
// ReadTimeout and WriteTimeout stay unset unless given, because they would cut
// off streaming responses and large uploads.
type Config struct {
	Addr    string
	Handler http.Handler

	// H2C also accepts HTTP/2 without TLS, using prior knowledge, as gRPC and
	// Connect clients do.
	H2C bool

	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration

	// TLSCertFile and TLSKeyFile make Run serve HTTPS when both are set.
	TLSCertFile string
	TLSKeyFile  string
}

// New returns a server for cfg without starting it.
func New(cfg Config) *http.Server {
	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           cfg.Handler,
		ReadHeaderTimeout: orDefault(cfg.ReadHeaderTimeout, defaultReadHeaderTimeout),
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       orDefault(cfg.IdleTimeout, defaultIdleTimeout),
	}

	if cfg.H2C {
		protocols := new(http.Protocols)
		protocols.SetHTTP1(true)
		protocols.SetHTTP2(true)
		protocols.SetUnencryptedHTTP2(true)
		srv.Protocols = protocols
	}

	return srv
}

// Run serves until ctx is cancelled and then shuts down gracefully. It returns
// nil after a clean stop, or the error if the server could not start or stop.
// Connections still open after cfg.ShutdownTimeout, such as event streams, are
// closed and logged rather than reported as an error.
func Run(ctx context.Context, srv *http.Server, cfg Config) error {
	served := make(chan error, 1)

	go func() {
		served <- serve(srv, cfg)
	}()

	select {
	case err := <-served:
		return ignoreClosed(err)
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), orDefault(cfg.ShutdownTimeout, defaultShutdownTimeout))
	defer cancel()

	err := srv.Shutdown(shutdownCtx)

	if errors.Is(err, context.DeadlineExceeded) {
		log.WithField("addr", srv.Addr).Warn("HTTP server shutdown timed out; closing remaining connections")
		err = srv.Close()
	}

	if err != nil {
		return fmt.Errorf("http server shutdown: %w", err)
	}

	return ignoreClosed(<-served)
}

// OK returns a handler that answers 200 with body as plain text, for liveness
// and readiness probes.
func OK(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write([]byte(body)); err != nil {
			log.WithError(err).Warn("Failed to write probe response")
		}
	}
}

func serve(srv *http.Server, cfg Config) error {
	if cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" {
		return srv.ListenAndServeTLS(cfg.TLSCertFile, cfg.TLSKeyFile)
	}

	return srv.ListenAndServe()
}

func ignoreClosed(err error) error {
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}

func orDefault(value, fallback time.Duration) time.Duration {
	if value == 0 {
		return fallback
	}

	return value
}
