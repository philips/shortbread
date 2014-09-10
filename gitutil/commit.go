package gitutil

import (
	"errors"
	"time"

	git "github.com/libgit2/git2go"
)

const (
	MasterRef = "refs/heads/master"
)

// AddAndCommit provides a convenient way of creating commits. Simply provide the (relative)paths of the files you want to stage as well as a commit message.
// equivalent to executing: git add /path/to/files.go and git commit -m "message"
func AddAndCommit(repo *git.Repository, paths []string, message string) error {
	if message == "" {
		return errors.New("commit message empty")
	}

	tree, err := gitAdd(repo, paths...)
	if err != nil {
		return err
	}
	defer tree.Free()

	err = gitCommit(repo, message, tree)
	return err
}

func gitCommit(repo *git.Repository, message string, tree *git.Tree) error {
	signature := &git.Signature{
		Name:  "shantanu",
		Email: "shantanu.joshi@coreos.com",
		When:  time.Now(),
	}
	parentCommit := make([]*git.Commit, 0)

	master, err := repo.LookupReference(MasterRef)
	if err != nil {
		_, err = repo.CreateCommit(MasterRef, signature, signature, message, tree, parentCommit...)
		return err
	}
	defer master.Free()

	objId := master.Target()
	c, err := repo.LookupCommit(objId)
	if err != nil {
		return err
	}

	parentCommit = append(parentCommit, c)
	_, err = repo.CreateCommit(MasterRef, signature, signature, message, tree, parentCommit...)
	return err
}
