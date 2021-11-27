package git

import (
	"fmt"
	"log"

	"github.com/go-git/go-git/v5"
	gitplumbing "github.com/go-git/go-git/v5/plumbing"
)

func GetLocalGitHeadHash(path string) (gitplumbing.Hash, error) {
	nilHash := gitplumbing.Hash{}

	repo, err := git.PlainOpen(path)
	if err != nil {
		return nilHash, fmt.Errorf("opening Git repository at %q: %w", path, err)
	}

	head, err := repo.Head()
	if err != nil {
		return nilHash, fmt.Errorf("obtaining Git HEAD reference on repository at %q: %w", path, err)
	}

	// Would prefer to use the short hash for brevity, but there does not seem to be support in go-git as yet.
	// See https://github.com/src-d/go-git/issues/602
	log.Printf("local hash is %q", head.Hash())
	return head.Hash(), nil
}

func GetRemoteHeadHash(localPath, branch string) (gitplumbing.Hash, error) {
	nilHash := gitplumbing.Hash{}

	repo, err := git.PlainOpen(localPath)
	if err != nil {
		return nilHash, fmt.Errorf("opening Git repository at %q: %w", localPath, err)
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		return nilHash, fmt.Errorf("obtaining remote on repository at %q: %w", localPath, err)
	}

	refs, err := remote.List(&git.ListOptions{})
	if err != nil {
		return nilHash, fmt.Errorf("listing remote refs for repository at %q: %w", localPath, err)
	}

	// Get the HEAD ref of the target branch, see https://github.com/src-d/go-git/issues/767
	targetRefName := "refs/heads/" + branch
	for _, ref := range refs {
		// Skip unwanted branch refs
		if ref.Name().String() != targetRefName {
			continue
		}

		// Target ref located
		log.Printf("remote hash is %q", ref.Hash())
		return ref.Hash(), nil
	}

	// If we get here there was a failure to locate the target branch ref
	return nilHash, fmt.Errorf("unable to locate target branch %q reference for repository at %q", branch, localPath)
}

func CloneGitRepo(URL, branch, destPath string) error {
	log.Printf("cloning branch %q of repository %q", branch, URL)

	// Options to clone just the master branch of the remote repo
	cloneOpts := &git.CloneOptions{
		URL:           URL,
		SingleBranch:  true,
		ReferenceName: gitplumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch)), // See https://github.com/src-d/go-git/issues/553
	}

	// Clone to local directory
	if _, err := git.PlainClone(destPath, false, cloneOpts); err != nil {
		return fmt.Errorf("cloning branch %q of repository %q: %w", branch, URL, err)
	}

	return nil
}

func PullGitRepo(path, branch string) error {
	log.Printf("pulling branch %q of repository at %q", branch, path)

	repo, err := git.PlainOpen(path)
	if err != nil {
		return fmt.Errorf("opening Git repository at %q: %w", path, err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("getting worktree for Git repository at %q: %w", path, err)
	}

	err = worktree.Pull(&git.PullOptions{
		RemoteName: "origin",
		SingleBranch: true,
		ReferenceName: gitplumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch)),
	})
	if err == git.NoErrAlreadyUpToDate {
		return nil
	} else if err != nil {
		return fmt.Errorf("pulling branch %q of Git repository at %q: %w", branch, path, err)
	}

	return nil
}
