package git

import (
	"fmt"
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
	Log bool
	Branch string
}

type GitShowHeadResult struct {
	Hash string
	AuthorName string
	AuthorEmail string
	AuthorDate string
	CommitterName string
	CommitterEmail string
	CommitterDate string
	Subject string
}

type GitStatusEntry struct {
	IndexStatus string
	WorkingTreeStatus string
	FilePath string
}

type GitStatusResult struct {
	Entries []GitStatusEntry
}

// getRemoteName attempts to get the remote name, defaulting to "origin" if not found
func getRemoteName(repoPath string) string {
	req := &easyexec.ExecRequest{
		Executable: "git",
		Args: []string{"remote"},
		WorkingDirectory: repoPath,
		Timeout: 10.0,
		Log: false,
	}
	result := easyexec.ExecWithRequest(req)
	if result.ExitCode == 0 && strings.TrimSpace(result.Output) != "" {
		// Get the first remote name
		remotes := strings.Fields(strings.TrimSpace(result.Output))
		if len(remotes) > 0 {
			return remotes[0]
		}
	}
	// Default to "origin" if no remotes found or command failed
	return "origin"
}

func CloneOrPull(req *CloneOrPullRequest) (*CloneOrPullResult) {
	repoName := path.Base(req.GitUrl)
	repoName = strings.TrimSuffix(repoName, ".git")

	if req.Branch == "" {
		req.Branch = "main"
	}

	log.WithFields(log.Fields{
		"gitUrl":    req.GitUrl,
		"localDir":  req.LocalDir,
		"repoName":  repoName,
		"branch":    req.Branch,
	}).Infof("GitPull")

	if req.Timeout < 60.0 {
		req.Timeout = 60.0 
	}

	if _, err := os.Stat(req.LocalDir); os.IsNotExist(err) {
		if err := os.Mkdir(req.LocalDir, 0755); err != nil {
			return &CloneOrPullResult{
				RepoName: repoName,
				WasCloned: false,
				ExecResult: &easyexec.ExecResult{
					ExitCode: 1,
					Error: fmt.Errorf("failed to create directory %s: %w", req.LocalDir, err),
				},
			}
		}
	}

	if _, err := os.Stat(filepath.Join(req.LocalDir, repoName)); os.IsNotExist(err) {
		req := &easyexec.ExecRequest{
			Executable: "git",
			Args: []string{"clone", "-b", req.Branch, req.GitUrl},
			WorkingDirectory: req.LocalDir,
			Timeout: req.Timeout, 
			Log: req.Log,
		}

		return &CloneOrPullResult{
			RepoName: repoName,
			WasCloned: true,
			ExecResult: easyexec.ExecWithRequest(req),
		}
	} else {
		if err != nil {
			log.Errorf("%v", err)
		}

		repoPath := filepath.Join(req.LocalDir, repoName)

		// Get the remote name (defaults to "origin" if not found)
		remoteName := getRemoteName(repoPath)

		// Checkout the branch first if not already on it
		checkoutReq := &easyexec.ExecRequest{
			Executable: "git",
			Args: []string{"checkout", req.Branch},
			WorkingDirectory: repoPath,
			Timeout: req.Timeout,
			Log: req.Log,
		}
		checkoutResult := easyexec.ExecWithRequest(checkoutReq)
		if checkoutResult.ExitCode != 0 {
			// If checkout fails, try to fetch the remote branch and create/checkout local branch
			fetchReq := &easyexec.ExecRequest{
				Executable: "git",
				Args: []string{"fetch", remoteName, req.Branch},
				WorkingDirectory: repoPath,
				Timeout: req.Timeout,
				Log: req.Log,
			}
			fetchResult := easyexec.ExecWithRequest(fetchReq)
			if fetchResult.ExitCode != 0 {
				// If fetch fails, return error result
				return &CloneOrPullResult{
					RepoName: repoName,
					WasCloned: false,
					ExecResult: fetchResult,
				}
			}

			// Try to checkout the branch (it might exist now after fetch)
			checkoutResult = easyexec.ExecWithRequest(checkoutReq)
			if checkoutResult.ExitCode != 0 {
				// If branch still doesn't exist locally, create it tracking the remote branch
				checkoutBranchReq := &easyexec.ExecRequest{
					Executable: "git",
					Args: []string{"checkout", "-b", req.Branch, remoteName + "/" + req.Branch},
					WorkingDirectory: repoPath,
					Timeout: req.Timeout,
					Log: req.Log,
				}
				checkoutResult = easyexec.ExecWithRequest(checkoutBranchReq)
				if checkoutResult.ExitCode != 0 {
					// If checkout still fails, return error result
					return &CloneOrPullResult{
						RepoName: repoName,
						WasCloned: false,
						ExecResult: checkoutResult,
					}
				}
			}
		}

		// Pull the latest changes
		pullReq := &easyexec.ExecRequest{
			Executable: "git",
			Args: []string{"pull", remoteName, req.Branch},
			WorkingDirectory: repoPath,
			Timeout: req.Timeout,
			Log: req.Log,
		}

		return &CloneOrPullResult{
			RepoName: repoName,
			WasCloned: false,
			ExecResult: easyexec.ExecWithRequest(pullReq),
		}
	}
}

func ShowHead(repoDir string) (*GitShowHeadResult, error) {
	log.WithFields(log.Fields{
		"repoDir": repoDir,
	}).Infof("GitShowHead")

	// Validate that repoDir exists and is a git repository
	if info, err := os.Stat(repoDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("repository directory does not exist: %s", repoDir)
	} else if err != nil {
		return nil, fmt.Errorf("failed to stat repository directory %s: %w", repoDir, err)
	} else if !info.IsDir() {
		return nil, fmt.Errorf("repository path is not a directory: %s", repoDir)
	}

	// Check if it's a git repository (.git can be a directory or a file in case of worktrees)
	gitDir := filepath.Join(repoDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("not a git repository: %s", repoDir)
	}

	req := &easyexec.ExecRequest{
		Executable: "git",
		Args: []string{"show", "-s", "--date=iso-strict", "--format=%H|%an|%ae|%ad|%cn|%ce|%cd|%s", "HEAD"},
		WorkingDirectory: repoDir,
		Timeout: 10.0,
		Log: false,
	}

	execResult := easyexec.ExecWithRequest(req)
	if execResult.ExitCode != 0 {
		return nil, execResult.Error
	}

	output := strings.TrimSpace(execResult.Output)
	parts := strings.Split(output, "|")
	
	if len(parts) != 8 {
		if execResult.Error != nil {
			return nil, execResult.Error
		}
		return nil, fmt.Errorf("unexpected output format from git show: expected 8 parts separated by '|', got %d parts. Output: %q", len(parts), output)
	}

	return &GitShowHeadResult{
		Hash:           parts[0],
		AuthorName:     parts[1],
		AuthorEmail:    parts[2],
		AuthorDate:     parts[3],
		CommitterName:  parts[4],
		CommitterEmail: parts[5],
		CommitterDate:  parts[6],
		Subject:        parts[7],
	}, nil
}

func Status(repoDir string) (*GitStatusResult, error) {
	log.WithFields(log.Fields{
		"repoDir": repoDir,
	}).Infof("GitStatus")

	// Validate that repoDir exists and is a git repository
	if info, err := os.Stat(repoDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("repository directory does not exist: %s", repoDir)
	} else if err != nil {
		return nil, fmt.Errorf("failed to stat repository directory %s: %w", repoDir, err)
	} else if !info.IsDir() {
		return nil, fmt.Errorf("repository path is not a directory: %s", repoDir)
	}

	// Check if it's a git repository (.git can be a directory or a file in case of worktrees)
	gitDir := filepath.Join(repoDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("not a git repository: %s", repoDir)
	}

	req := &easyexec.ExecRequest{
		Executable: "git",
		Args: []string{"status", "--porcelain"},
		WorkingDirectory: repoDir,
		Timeout: 10.0,
		Log: false,
	}

	execResult := easyexec.ExecWithRequest(req)
	if execResult.ExitCode != 0 {
		return nil, execResult.Error
	}

	output := strings.TrimSpace(execResult.Output)
	if output == "" {
		return &GitStatusResult{Entries: []GitStatusEntry{}}, nil
	}

	lines := strings.Split(output, "\n")
	entries := make([]GitStatusEntry, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if len(line) < 3 {
			continue
		}

		indexStatus := string(line[0])
		workingTreeStatus := string(line[1])
		filePath := strings.TrimSpace(line[2:])

		// For renamed/copied files, git status --porcelain shows "R  old -> new" or "C  old -> new"
		// After trimming, filePath contains "old -> new" which we preserve as-is

		entries = append(entries, GitStatusEntry{
			IndexStatus:      indexStatus,
			WorkingTreeStatus: workingTreeStatus,
			FilePath:         filePath,
		})
	}

	return &GitStatusResult{Entries: entries}, nil
}

