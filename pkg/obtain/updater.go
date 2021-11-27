package obtain

import (
	"fmt"

	gitutil "github.com/jhwbarlow/mockcicd/pkg/git"
)

type Updater interface {
	Update(path string) error
}

type GitPullUpdater struct {
	Branch string
}

func NewGitPullUpdater(branch string) *GitPullUpdater {
	return &GitPullUpdater{
		Branch: branch,
	}
}

func (u *GitPullUpdater) Update(path string) error {
	if err := gitutil.PullGitRepo(path, u.Branch); err != nil {
		return fmt.Errorf("pulling Git repo: %w", err)
	}

	return nil
}
