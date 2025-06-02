package git

import (
	log "github.com/sirupsen/logrus"
	"github.com/jamesread/golure/pkg/easyexec"
	"os"
	"path"
	"path/filepath"
)

func CloneOrPull(gitUrl string, localDir string) (*easyexec.ExecResult) {
	repoName := path.Base(gitUrl)

	log.WithFields(log.Fields{
		"gitUrl":    gitUrl,
		"localDir":  localDir,
		"repoName":  repoName,
	}).Infof("GitPull")

	if _, err := os.Stat(localDir); os.IsNotExist(err) {
		os.Mkdir(localDir, 0755)
	}

	if _, err := os.Stat(filepath.Join(localDir, repoName)); os.IsNotExist(err) {
		err = os.Chdir(localDir)

		if err != nil {
			log.Errorf("%v", err)
		}

		return easyexec.ExecLog("git", []string {"clone", gitUrl})
	} else {
		err = os.Chdir(filepath.Join(localDir, repoName))

		if err != nil {
			log.Errorf("%v", err)
		}

		return easyexec.ExecLog("git", []string {"pull"})
	}
}

