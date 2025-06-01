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
		
		log.Debugf("Looking for %v directory at path: %s %v\n", name, abspath)

		if err != nil {
			continue
		}

		if stat.IsDir() {
			return abspath, nil
		}
	}

	return "", errors.New("No existing directory found in the provided list")
}
