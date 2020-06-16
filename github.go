package covid19brazilimporter

import (
	"fmt"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func cloneRepo(folder, password, branch string) (*git.Repository, error) {
	url := fmt.Sprintf("https://luiz290788:%s@github.com/luiz290788/covid19-brazil.git", password)
	return git.PlainClone(folder, false, &git.CloneOptions{
		URL:           url,
		Depth:         1,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
	})

}

func commitAll(repository *git.Repository, message, name, email string) (hash string, err error) {
	var worktree *git.Worktree
	worktree, err = repository.Worktree()
	if err != nil {
		return
	}

	err = worktree.AddGlob("src/data/covid/*.json")
	if err != nil {
		return
	}

	var commitHash plumbing.Hash
	commitHash, err = worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Email: email,
			Name:  name,
			When:  time.Now(),
		},
		Committer: &object.Signature{
			Email: email,
			Name:  name,
			When:  time.Now(),
		},
	})

	if err != nil {
		return
	}

	return commitHash.String(), nil
}
