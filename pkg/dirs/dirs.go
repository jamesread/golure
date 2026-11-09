package dirs

import (
	"errors"

	log "github.com/sirupsen/logrus"
)

// GetFirstExistingDirectory returns the first of directories that exists. Use
// FindDirectories for more control and information.
func GetFirstExistingDirectory(name string, directories []string) (string, error) {
	existing := FindDirectories(Search{Name: name, Paths: directories}).Existing()

	if len(existing) > 0 {
		log.Debugf("Found %v directory at path: %s", name, existing[0].Path)
		return existing[0].Path, nil
	}

	log.Warnf("No existing %v directory found in the provided list", name)

	return "not-found", errors.New("No existing directory found in the provided list")
}

// GetFirstExistingFileFromDirs returns the path of filename in the first of
// directories that contains it. Use FindFiles for more control and
// information.
func GetFirstExistingFileFromDirs(name string, directories []string, filename string) (string, error) {
	existing := FindFiles(Search{Name: name, Filenames: []string{filename}, Paths: directories}).Existing()

	if len(existing) > 0 {
		log.Debugf("Found %v file at path: %s", name, existing[0].Path)
		return existing[0].Path, nil
	}

	log.Warnf("No existing %v file found in the provided list of directories", name)

	return "not-found", errors.New("No existing file found in the provided list of directories")
}
