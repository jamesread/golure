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
	Timeout float64
}

type CloneOrPullRequest struct {
	GitUrl string
	LocalDir string
	Timeout float64
}

func CloneOrPull(req *CloneOrPullRequest) (*CloneOrPullResult) {
	repoName := path.Base(gitUrl)
	repoName = strings.TrimSuffix(repoName, ".git")

	log.WithFields(log.Fields{
		"gitUrl":    req.GitUrl,
		"localDir":  req.LocalDir,
		"repoName":  repoName,
	}).Infof("GitPull")

	if req.Timeout <= 0 {
		req.Timeout = 60.0 // Default timeout of 60 seconds
	}

	if _, err := os.Stat(req.LocalDir); os.IsNotExist(err) {
		os.Mkdir(req.LocalDir, 0755)
	}

	if _, err := os.Stat(filepath.Join(req.LocalDir, repoName)); os.IsNotExist(err) {
		req := &easyexec.ExecRequest{
			Executable: "git",
			Args: []string{"clone", req.GitUrl},
			WorkingDirectory: req.LocalDir,
			Timeout: req.Timeout, 
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
			WorkingDirectory: filepath.Join(req.LocalDir, repoName),
		}

		return &CloneOrPullResult{
			RepoName: repoName,
			WasCloned: false,
			ExecResult: easyexec.ExecWithReqLog(req),
		}
	}
}

