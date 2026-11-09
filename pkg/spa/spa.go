// Package spa serves a built single-page application from a directory.
package spa

import (
	"net/http"
	"os"
	"path"
	"strings"
)

const noCacheValue = "no-store, no-cache, must-revalidate, max-age=0"

// Options adjusts how Handler responds. The zero value is ready to use.
type Options struct {
	// Exclude lists exact URL paths that return 404 rather than falling back
	// to index.html when no file exists there.
	Exclude []string

	// NoCache reports whether a response for the cleaned URL path must not be
	// cached, for example a service worker script. Nil means none.
	NoCache func(urlPath string) bool
}

// Handler serves files from root. A path with no matching file gets
// index.html if it has no file extension, so client-side routes work on
// refresh, and 404 otherwise, so a missing asset is not answered with HTML.
// Only GET and HEAD are served. Files are opened through os.Root, so neither
// the URL nor a symlink can reach outside root. Root is opened per request, so
// it may be created or rebuilt while the server runs.
func Handler(root string, opts Options) http.Handler {
	excluded := make(map[string]struct{}, len(opts.Exclude))

	for _, p := range opts.Exclude {
		excluded[p] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}

		// The leading slash matters: a handler mounted behind
		// http.StripPrefix("/", ...) receives paths without one, and Clean
		// only removes ".." elements from rooted paths.
		urlPath := path.Clean("/" + r.URL.Path)

		if opts.NoCache != nil && opts.NoCache(urlPath) {
			w.Header().Set("Cache-Control", noCacheValue)
		}

		files, err := os.OpenRoot(root)

		if err != nil {
			http.NotFound(w, r)
			return
		}

		defer func() { _ = files.Close() }()

		if serve(w, r, files, urlPath) {
			return
		}

		_, isExcluded := excluded[urlPath]

		if isExcluded || path.Ext(urlPath) != "" || !serve(w, r, files, "/index.html") {
			http.NotFound(w, r)
		}
	})
}

// serve writes the file at name, or the index.html inside it if name is a
// directory, and reports whether it did. It uses http.ServeContent because
// http.FileServer redirects any path ending in /index.html.
func serve(w http.ResponseWriter, r *http.Request, files *os.Root, name string) bool {
	relative := strings.TrimPrefix(name, "/")

	if relative == "" {
		relative = "."
	}

	file, err := files.Open(relative)

	if err != nil {
		return false
	}

	defer func() { _ = file.Close() }()

	info, err := file.Stat()

	if err != nil {
		return false
	}

	if info.IsDir() {
		if path.Base(name) == "index.html" {
			return false
		}

		return serve(w, r, files, path.Join(name, "index.html"))
	}

	if !info.Mode().IsRegular() {
		return false
	}

	http.ServeContent(w, r, info.Name(), info.ModTime(), file)

	return true
}
