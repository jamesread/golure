package spa

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const indexBody = "<html>index</html>"

func newRoot(t *testing.T) string {
	t.Helper()

	parent := t.TempDir()
	root := filepath.Join(parent, "dist")

	require.NoError(t, os.MkdirAll(filepath.Join(root, "assets"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "index.html"), []byte(indexBody), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "assets", "app.js"), []byte("js"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "sw.js"), []byte("sw"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(parent, "secret.txt"), []byte("secret"), 0o644))

	return root
}

func get(h http.Handler, method, target string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/", nil)
	req.URL.Path = target

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	return rec
}

func TestHandler(t *testing.T) {
	h := Handler(newRoot(t), Options{Exclude: []string{"/api"}})

	tests := []struct {
		name   string
		method string
		path   string
		status int
		body   string
	}{
		{"root", http.MethodGet, "/", http.StatusOK, indexBody},
		{"index without redirect", http.MethodGet, "/index.html", http.StatusOK, indexBody},
		{"deep link without redirect", http.MethodGet, "/orders/42", http.StatusOK, indexBody},
		{"real asset", http.MethodGet, "/assets/app.js", http.StatusOK, "js"},
		{"directory without index", http.MethodGet, "/assets", http.StatusOK, indexBody},
		{"missing asset", http.MethodGet, "/assets/gone.js", http.StatusNotFound, ""},
		{"excluded path", http.MethodGet, "/api", http.StatusNotFound, ""},
		{"no leading slash", http.MethodGet, "assets/app.js", http.StatusOK, "js"},
		{"traversal stays inside root", http.MethodGet, "/../secret.txt", http.StatusNotFound, ""},
		{"traversal without leading slash", http.MethodGet, "../secret.txt", http.StatusNotFound, ""},
		{"post rejected", http.MethodPost, "/", http.StatusNotFound, ""},
		{"head allowed", http.MethodHead, "/orders", http.StatusOK, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := get(h, tt.method, tt.path)

			assert.Equal(t, tt.status, rec.Code)
			assert.Empty(t, rec.Header().Get("Location"))

			if tt.body != "" {
				assert.Equal(t, tt.body, rec.Body.String())
			}

			assert.NotContains(t, rec.Body.String(), "secret")
		})
	}
}

func TestHandlerRejectsSymlinkOutsideRoot(t *testing.T) {
	root := newRoot(t)

	require.NoError(t, os.Symlink(filepath.Join(filepath.Dir(root), "secret.txt"), filepath.Join(root, "leak.txt")))

	rec := get(Handler(root, Options{}), http.MethodGet, "/leak.txt")

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.NotContains(t, rec.Body.String(), "secret")
}

func TestHandlerNoCache(t *testing.T) {
	h := Handler(newRoot(t), Options{NoCache: func(p string) bool {
		return strings.HasSuffix(p, "/sw.js")
	}})

	assert.Contains(t, get(h, http.MethodGet, "/sw.js").Header().Get("Cache-Control"), "no-store")
	assert.Empty(t, get(h, http.MethodGet, "/assets/app.js").Header().Get("Cache-Control"))
}

func TestHandlerMissingRoot(t *testing.T) {
	h := Handler(filepath.Join(t.TempDir(), "absent"), Options{})

	assert.Equal(t, http.StatusNotFound, get(h, http.MethodGet, "/").Code)
}

func TestHandlerWithoutIndex(t *testing.T) {
	h := Handler(t.TempDir(), Options{})

	assert.Equal(t, http.StatusNotFound, get(h, http.MethodGet, "/orders").Code)
}
