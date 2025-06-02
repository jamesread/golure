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
		if err != nil {
			log.Errorf("%v", err)
		}

		req := &easyexec.ExecRequest{
			Executable: "git",
			Args: []string{"clone", gitUrl},
			WorkingDirectory: localDir,
		}

		return easyexec.ExecWithRequest(req)
	} else {
		if err != nil {
			log.Errorf("%v", err)
		}

		req := &easyexec.ExecRequest{
			Executable: "git",
			Args: []string{"pull"},
			WorkingDirectory: filepath.Join(localDir, repoName),
		}

		return easyexec.ExecWithReqLog(req)
	}
}

