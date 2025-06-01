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
		
		log.Debugf("Looking for %v directory at path: %s\n", name, abspath)

		if err != nil {
			continue
		}

		if stat.IsDir() {
			log.Debugf("Found %v directory at path: %s\n", name, abspath)
			return abspath, nil
		}
	}

	log.Warnf("No existing %v directory found in the provided list", name)

	return "not-found", errors.New("No existing directory found in the provided list")
}
