package dirs

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Search describes where to look and what to look for. Always use keyed
// fields: new criteria may be added, and their zero value will preserve the
// existing behaviour.
type Search struct {
	// Name is a label for what is being searched for, used in log messages.
	Name string

	// Filenames are the allowable names within each of Paths, most preferred
	// first. If empty, each of Paths is evaluated directly.
	Filenames []string

	// Paths are the locations to search, most preferred first. A leading ~,
	// $HOME or ${HOME} is expanded and relative paths are resolved against
	// the working directory. Empty entries are ignored, so an unset
	// environment variable does not mean the working directory; use "." for
	// that.
	Paths []string
}

// Candidate is a single evaluated path. New information may be added as new
// fields.
type Candidate struct {
	// Path is always absolute.
	Path string

	// Exists is true if something of the kind searched for is at Path. A
	// directory found by FindFiles, or a file found by FindDirectories, does
	// not count.
	Exists bool

	// Readable is true if Exists and Path could be opened for reading. Only
	// regular files and directories are opened, because opening others, such
	// as a FIFO, can block.
	Readable bool

	// Err is why Path could not be examined or opened, such as a permission
	// error. It is nil if Path simply does not exist.
	Err error
}

// Result holds every evaluated path, in preference order: all of Filenames in
// the first of Paths, then all of Filenames in the second of Paths, and so
// on. New information may be added as new fields.
type Result struct {
	// Candidates excludes only paths that could not be made absolute, for
	// example if the home directory is unknown. These are logged as warnings.
	Candidates []Candidate
}

// Existing returns the candidates that exist, in preference order.
func (r Result) Existing() []Candidate {
	return r.filter(func(c Candidate) bool { return c.Exists })
}

// Readable returns the candidates that exist and are readable, in preference
// order.
func (r Result) Readable() []Candidate {
	return r.filter(func(c Candidate) bool { return c.Readable })
}

// First returns the most preferred readable candidate, and false if there is
// none.
func (r Result) First() (Candidate, bool) {
	for _, c := range r.Candidates {
		if c.Readable {
			return c, true
		}
	}

	return Candidate{}, false
}

func (r Result) filter(keep func(Candidate) bool) []Candidate {
	var kept []Candidate

	for _, c := range r.Candidates {
		if keep(c) {
			kept = append(kept, c)
		}
	}

	return kept
}

// FindFiles evaluates every combination of search.Paths and search.Filenames
// as a file. Not finding anything is not an error; inspect the Result.
func FindFiles(search Search) Result {
	return find(search, false)
}

// FindDirectories evaluates every combination of search.Paths and
// search.Filenames as a directory. Not finding anything is not an error;
// inspect the Result.
func FindDirectories(search Search) Result {
	return find(search, true)
}

func find(search Search, wantDir bool) Result {
	filenames := search.Filenames

	if len(filenames) == 0 {
		filenames = []string{""}
	}

	var result Result

	for _, path := range search.Paths {
		if path == "" {
			continue
		}

		base, err := expandHome(path)

		if err != nil {
			log.Warnf("Skipping %v search path %s: %v", search.Name, path, err)
			continue
		}

		for _, filename := range filenames {
			abspath, err := filepath.Abs(filepath.Join(base, filename))

			if err != nil {
				log.Warnf("Skipping %v search path %s: %v", search.Name, path, err)
				continue
			}

			candidate := inspect(abspath, wantDir)

			log.Debugf("Evaluated %v at path: %s (exists: %v, readable: %v)", search.Name, abspath, candidate.Exists, candidate.Readable)

			result.Candidates = append(result.Candidates, candidate)
		}
	}

	return result
}

func inspect(path string, wantDir bool) Candidate {
	candidate := Candidate{Path: path}

	info, err := os.Stat(path)

	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			candidate.Err = err
		}

		return candidate
	}

	if info.IsDir() != wantDir {
		return candidate
	}

	candidate.Exists = true

	if !info.Mode().IsRegular() && !info.IsDir() {
		return candidate
	}

	handle, err := os.Open(path)

	if err != nil {
		candidate.Err = err
		return candidate
	}

	handle.Close()

	candidate.Readable = true

	return candidate
}

// expandHome replaces a leading ~, $HOME or ${HOME} path component with the
// user's home directory. Anything else, such as ~other or $HOMEPAGE, is
// returned unchanged.
func expandHome(path string) (string, error) {
	for _, prefix := range []string{"~", "$HOME", "${HOME}"} {
		rest, found := strings.CutPrefix(path, prefix)

		if !found {
			continue
		}

		if rest != "" && !os.IsPathSeparator(rest[0]) {
			return path, nil
		}

		home, err := os.UserHomeDir()

		if err != nil {
			return "", err
		}

		return home + rest, nil
	}

	return path, nil
}
