package check

import (
	"fmt"

	gitutil "github.com/jhwbarlow/mockcicd/pkg/git"
)

type Checker interface {
	Check() (bool, error)
}

type GitChecker struct {
	Path         string
	RemoteBranch string
}

func NewGitChecker(path, remoteBranch string) *GitChecker {
	return &GitChecker{
		Path:         path,
		RemoteBranch: remoteBranch,
	}
}

func (c *GitChecker) Check() (bool, error) {
	localHash, err := gitutil.GetLocalGitHeadHash(c.Path)
	if err != nil {
		return false, fmt.Errorf("getting local git hash: %w", err)
	}

	remoteHash, err := gitutil.GetRemoteHeadHash(c.Path, c.RemoteBranch)
	if err != nil {
		return false, fmt.Errorf("getting remote git hash: %w", err)
	}

	return localHash != remoteHash, nil
}
