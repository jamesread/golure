package httpserve

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func freeAddr(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	defer ln.Close()

	return ln.Addr().String()
}

func waitForListen(t *testing.T, addr string) {
	t.Helper()

	require.Eventually(t, func() bool {
		conn, err := net.Dial("tcp", addr)

		if err != nil {
			return false
		}

		conn.Close()

		return true
	}, 5*time.Second, 10*time.Millisecond)
}

func start(t *testing.T, cfg Config) (*http.Server, context.CancelFunc, <-chan error) {
	t.Helper()

	cfg.Addr = freeAddr(t)
	srv := New(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	done := make(chan error, 1)

	go func() {
		done <- Run(ctx, srv, cfg)
	}()

	waitForListen(t, cfg.Addr)

	return srv, cancel, done
}

func receive(t *testing.T, done <-chan error) error {
	t.Helper()

	select {
	case err := <-done:
		return err
	case <-time.After(5 * time.Second):
		require.Fail(t, "Run did not return")
		return nil
	}
}

func TestNewDefaults(t *testing.T) {
	srv := New(Config{Addr: ":1"})

	assert.Equal(t, 10*time.Second, srv.ReadHeaderTimeout)
	assert.Equal(t, 120*time.Second, srv.IdleTimeout)
	assert.Zero(t, srv.ReadTimeout)
	assert.Zero(t, srv.WriteTimeout)
	assert.Nil(t, srv.Protocols)
}

func TestNewKeepsExplicitTimeouts(t *testing.T) {
	srv := New(Config{ReadHeaderTimeout: time.Second, ReadTimeout: 2 * time.Second, WriteTimeout: 3 * time.Second, IdleTimeout: 4 * time.Second})

	assert.Equal(t, time.Second, srv.ReadHeaderTimeout)
	assert.Equal(t, 2*time.Second, srv.ReadTimeout)
	assert.Equal(t, 3*time.Second, srv.WriteTimeout)
	assert.Equal(t, 4*time.Second, srv.IdleTimeout)
}

func TestH2CRequest(t *testing.T) {
	srv, cancel, done := start(t, Config{H2C: true, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, r.Proto)
	})})

	protocols := new(http.Protocols)
	protocols.SetUnencryptedHTTP2(true)

	client := &http.Client{Transport: &http.Transport{Protocols: protocols, TLSClientConfig: &tls.Config{}}}

	resp, err := client.Get("http://" + srv.Addr + "/")
	require.NoError(t, err)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "HTTP/2.0", string(body))

	cancel()
	assert.NoError(t, receive(t, done))
}

func TestRunReturnsNilOnCancel(t *testing.T) {
	_, cancel, done := start(t, Config{Handler: OK("ok")})

	cancel()

	assert.NoError(t, receive(t, done))
}

func TestRunCompletesInFlightRequest(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})

	srv, cancel, done := start(t, Config{Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(entered)
		<-release
		_, _ = io.WriteString(w, "finished")
	})})

	responses := make(chan string, 1)

	go func() {
		resp, err := http.Get("http://" + srv.Addr + "/")

		if err != nil {
			responses <- err.Error()
			return
		}

		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		responses <- string(body)
	}()

	<-entered
	cancel()

	select {
	case err := <-done:
		require.Fail(t, "Run returned before the request finished", "%v", err)
	case <-time.After(100 * time.Millisecond):
	}

	close(release)

	assert.Equal(t, "finished", <-responses)
	assert.NoError(t, receive(t, done))
}

func TestRunForceClosesAfterShutdownTimeout(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })

	srv, cancel, done := start(t, Config{ShutdownTimeout: 50 * time.Millisecond, Handler: http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		close(entered)
		<-release
	})})

	go func() {
		if resp, err := http.Get("http://" + srv.Addr + "/"); err == nil {
			resp.Body.Close()
		}
	}()

	<-entered
	cancel()

	assert.NoError(t, receive(t, done))
}

func TestRunReportsListenFailure(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	defer ln.Close()

	cfg := Config{Addr: ln.Addr().String(), Handler: OK("ok")}

	assert.Error(t, Run(context.Background(), New(cfg), cfg))
}

func TestOK(t *testing.T) {
	rec := httptest.NewRecorder()

	OK("healthy")(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "healthy", rec.Body.String())
	assert.Contains(t, rec.Header().Get("Content-Type"), "text/plain")
}
