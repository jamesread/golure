package git

import (
	log "github.com/sirupsen/logrus"
	"github.com/jamesread/golure/pkg/easyexec"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type CloneOrPullResult struct {
	RepoName string
	WasCloned bool
	ExecResult *easyexec.ExecResult
}

func CloneOrPull(gitUrl string, localDir string) (*CloneOrPullResult) {
	repoName := path.Base(gitUrl)
	repoName = strings.TrimSuffix(repoName, ".git")

	log.WithFields(log.Fields{
		"gitUrl":    gitUrl,
		"localDir":  localDir,
		"repoName":  repoName,
	}).Infof("GitPull")

	if _, err := os.Stat(localDir); os.IsNotExist(err) {
		os.Mkdir(localDir, 0755)
	}

	if _, err := os.Stat(filepath.Join(localDir, repoName)); os.IsNotExist(err) {
		req := &easyexec.ExecRequest{
			Executable: "git",
			Args: []string{"clone", gitUrl},
			WorkingDirectory: localDir,
		}

		return &CloneOrPullResult{
			RepoName: repoName,
			WasCloned: true,
			ExecResult: easyexec.ExecWithReqLog(req),
		}
	} else {
		if err != nil {
			log.Errorf("%v", err)
		}

		req := &easyexec.ExecRequest{
			Executable: "git",
			Args: []string{"pull"},
			WorkingDirectory: filepath.Join(localDir, repoName),
		}

		return &CloneOrPullResult{
			RepoName: repoName,
			WasCloned: false,
			ExecResult: easyexec.ExecWithReqLog(req),
		}
	}
}

