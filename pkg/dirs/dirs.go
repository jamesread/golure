package dirs

import (
	"os"
	"path/filepath"
	"errors"
	"strings"
	"fmt"
)

func GetFirstExistingDirectory(directories []string) (string, error) {
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
		
		fmt.Printf("Checking directory: %s %v\n", abspath, stat)

		if err != nil {
			continue
		}

		if stat.IsDir() {
			return abspath, nil
		}
	}

	return "", errors.New("No existing directory found in the provided list")
}
