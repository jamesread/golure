package dirs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, path string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte("x"), 0o644))
}

func paths(candidates []Candidate) []string {
	var result []string

	for _, c := range candidates {
		result = append(result, c.Path)
	}

	return result
}

func TestFindFilesReturnsAllCandidatesInOrder(t *testing.T) {
	first := t.TempDir()
	second := t.TempDir()

	writeFile(t, filepath.Join(first, "config.yaml"))
	writeFile(t, filepath.Join(second, "config.yml"))
	writeFile(t, filepath.Join(second, "config.yaml"))

	result := FindFiles(Search{
		Filenames: []string{"config.yml", "config.yaml"},
		Paths:     []string{first, second},
	})

	assert.Equal(t, []Candidate{
		{Path: filepath.Join(first, "config.yml")},
		{Path: filepath.Join(first, "config.yaml"), Exists: true, Readable: true},
		{Path: filepath.Join(second, "config.yml"), Exists: true, Readable: true},
		{Path: filepath.Join(second, "config.yaml"), Exists: true, Readable: true},
	}, result.Candidates)

	expected := []string{
		filepath.Join(first, "config.yaml"),
		filepath.Join(second, "config.yml"),
		filepath.Join(second, "config.yaml"),
	}

	assert.Equal(t, expected, paths(result.Existing()))
	assert.Equal(t, expected, paths(result.Readable()))

	best, found := result.First()

	assert.True(t, found)
	assert.Equal(t, expected[0], best.Path)
}

func TestFirstWithNothingFound(t *testing.T) {
	result := FindFiles(Search{
		Filenames: []string{"config.yml"},
		Paths:     []string{t.TempDir()},
	})

	_, found := result.First()

	assert.False(t, found)
	assert.Len(t, result.Candidates, 1)
	assert.Empty(t, result.Existing())
	assert.Empty(t, result.Readable())
}

func TestFindExpandsHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	writeFile(t, filepath.Join(home, "app", "config.yml"))

	expected := filepath.Join(home, "app", "config.yml")

	for _, path := range []string{"~/app", "$HOME/app", "${HOME}/app"} {
		result := FindFiles(Search{Filenames: []string{"config.yml"}, Paths: []string{path}})

		require.Len(t, result.Candidates, 1, path)
		assert.Equal(t, expected, result.Candidates[0].Path, path)
		assert.True(t, result.Candidates[0].Exists, path)
	}

	for _, path := range []string{"~", "$HOME", "${HOME}"} {
		result := FindDirectories(Search{Paths: []string{path}})

		require.Len(t, result.Candidates, 1, path)
		assert.Equal(t, home, result.Candidates[0].Path, path)
		assert.True(t, result.Candidates[0].Exists, path)
	}
}

func TestFindDoesNotExpandLookalikes(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	wd, err := os.Getwd()
	require.NoError(t, err)

	for _, path := range []string{"~other/app", "$HOMEPAGE/app", "app~/x"} {
		result := FindDirectories(Search{Paths: []string{path}})

		require.Len(t, result.Candidates, 1, path)
		assert.Equal(t, filepath.Join(wd, path), result.Candidates[0].Path, path)
	}
}

func TestFindResolvesRelativePaths(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)

	writeFile(t, filepath.Join(dir, "etc", "config.yml"))

	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(filepath.Join(dir, "etc")))

	t.Cleanup(func() { require.NoError(t, os.Chdir(wd)) })

	result := FindFiles(Search{
		Filenames: []string{"config.yml"},
		Paths:     []string{".", "../etc", ""},
	})

	expected := filepath.Join(dir, "etc", "config.yml")

	assert.Equal(t, []string{expected, expected}, paths(result.Readable()))
}

func TestFindRequiresMatchingKind(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "file"))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "subdir"), 0o755))

	search := Search{Filenames: []string{"file", "subdir"}, Paths: []string{dir}}

	assert.Equal(t, []string{filepath.Join(dir, "file")}, paths(FindFiles(search).Existing()))
	assert.Equal(t, []string{filepath.Join(dir, "subdir")}, paths(FindDirectories(search).Existing()))
}

func TestFindFilesWithoutFilenamesEvaluatesPathsDirectly(t *testing.T) {
	file := filepath.Join(t.TempDir(), "config.yml")

	writeFile(t, file)

	assert.Equal(t, []string{file}, paths(FindFiles(Search{Paths: []string{file}}).Readable()))
}

func TestFindSeparatesExistingFromReadable(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can read everything")
	}

	dir := t.TempDir()
	unreadable := filepath.Join(dir, "a.yml")
	readable := filepath.Join(dir, "b.yml")

	writeFile(t, unreadable)
	writeFile(t, readable)
	require.NoError(t, os.Chmod(unreadable, 0o000))

	result := FindFiles(Search{Filenames: []string{"a.yml", "b.yml"}, Paths: []string{dir}})

	assert.Equal(t, []string{unreadable, readable}, paths(result.Existing()))
	assert.Equal(t, []string{readable}, paths(result.Readable()))

	assert.True(t, result.Candidates[0].Exists)
	assert.False(t, result.Candidates[0].Readable)
	assert.ErrorIs(t, result.Candidates[0].Err, os.ErrPermission)
}

func TestGetFirstExistingFileFromDirs(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "config.yml")

	writeFile(t, file)

	found, err := GetFirstExistingFileFromDirs("test", []string{"/foo/bar", dir}, "config.yml")

	assert.NoError(t, err)
	assert.Equal(t, file, found)

	found, err = GetFirstExistingFileFromDirs("test", []string{"/foo/bar"}, "config.yml")

	assert.Error(t, err)
	assert.Equal(t, "not-found", found)
}
