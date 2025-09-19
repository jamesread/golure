package dirs

import (
	"os"
	"path/filepath"
	"errors"
	"strings"

	log "github.com/sirupsen/logrus"
)

func GetFirstExistingDirectory(name string, directories []string) (string, error) {
	for _, dir := range directories {
		if strings.Contains(dir, "~") {
			home, _ := os.UserHomeDir()
			dir = strings.ReplaceAll(dir, "~", home)
		}

		abspath, err := filepath.Abs(dir)
	
		if err != nil {
			continue
		}

		stat, err := os.Stat(dir)
		
		log.Debugf("Looking for %v directory at path: %s", name, abspath)

		if err != nil {
			continue
		}

		if stat.IsDir() {
			log.Debugf("Found %v directory at path: %s", name, abspath)
			return abspath, nil
		}
	}

	log.Warnf("No existing %v directory found in the provided list", name)

	return "not-found", errors.New("No existing directory found in the provided list")
}

func GetFirstExistingFileFromDirs(name string, directories []string, filename string) (string, error) {
	for _, dir := range directories {
		if strings.Contains(dir, "~") {
			home, _ := os.UserHomeDir()
			dir = strings.ReplaceAll(dir, "~", home)
		}

		abspath, err := filepath.Abs(filepath.Join(dir, filename))
	
		if err != nil {
			continue
		}

		stat, err := os.Stat(abspath)
		
		log.Debugf("Looking for %v file at path: %s", name, abspath)

		if err != nil {
			continue
		}

		if !stat.IsDir() {
			log.Debugf("Found %v file at path: %s", name, abspath)
			return abspath, nil
		}
	}

	log.Warnf("No existing %v file found in the provided list of directories", name)

	return "not-found", errors.New("No existing file found in the provided list of directories")
}
