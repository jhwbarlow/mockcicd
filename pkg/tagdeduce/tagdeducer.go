package tagdeduce

import (
	"fmt"

	gitutil "github.com/jhwbarlow/mockcicd/pkg/git"
)

type TagDeducer interface {
	Deduce() (string, error)
}

type GitHashTagDeducer struct {
	Path string
}

func NewGitHashTagDeducer(path string) *GitHashTagDeducer {
	return &GitHashTagDeducer{
		Path: path,
	}
}

func (d *GitHashTagDeducer) Deduce() (string, error) {
	hash, err := gitutil.GetLocalGitHeadHash(d.Path)
	if err != nil {
		return "", fmt.Errorf("getting local git hash: %w", err)
	}

	return hash.String(), nil
}
