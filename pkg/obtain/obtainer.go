package obtain

import (
	"fmt"

	gitutil "github.com/jhwbarlow/mockcicd/pkg/git"
	"github.com/jhwbarlow/mockcicd/pkg/prepare"
)

type Obtainer interface {
	Obtain(destPath string) error
}

type GitCloneObtainer struct {
	URL      string
	Branch   string
	Preparer prepare.Preparer
}

func NewGitCloneObtainer(URL, branch string, preparer prepare.Preparer) *GitCloneObtainer {
	return &GitCloneObtainer{
		URL:      URL,
		Branch:   branch,
		Preparer: preparer,
	}
}

func (o *GitCloneObtainer) Obtain(destPath string) error {
	if err := o.Preparer.Prepare(destPath); err != nil {
		return fmt.Errorf("preparing filesystem: %w", err)
	}

	if err := gitutil.CloneGitRepo(o.URL, o.Branch, destPath); err != nil {
		return fmt.Errorf("cloning Git repo: %w", err)
	}

	return nil
}
