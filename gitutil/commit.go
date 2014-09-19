package gitutil

import (
	"errors"
	"time"

	git "github.com/coreos/shortbread/Godeps/_workspace/src/github.com/libgit2/git2go"
)

const (
	MasterRef = "refs/heads/master"
)

// AddAndCommit provides a convenient way of creating commits. Simply provide the (relative)paths of the files you want to stage as well as a commit message.
// equivalent to executing: git add /path/to/files.go and git commit -m "message"
func AddAndCommit(repo *git.Repository, paths []string, message, authorName, authorEmail string) error {
	if message == "" {
		return errors.New("commit message empty")
	}

	tree, err := gitAdd(repo, paths...)
	if err != nil {
		return err
	}
	defer tree.Free()

	err = gitCommit(repo, message, tree, authorName, authorEmail)
	return err
}

func gitCommit(repo *git.Repository, message string, tree *git.Tree, authorName, authorEmail string) error {
	commitSignature := &git.Signature{
		Name:  "shortbread",
		Email: "shortbread@example.com",
		When:  time.Now(),
	}

	if authorName == "" {
		authorName = commitSignature.Name
	}

	if authorEmail == "" {
		authorEmail = commitSignature.Email
	}

	signature := &git.Signature{
		Name:  authorName,
		Email: authorEmail,
		When:  time.Now(),
	}

	parentCommit := make([]*git.Commit, 0)
	master, err := repo.LookupReference(MasterRef)
	if err != nil {
		_, err = repo.CreateCommit(MasterRef, signature, commitSignature, message, tree, parentCommit...)
		return err
	}
	defer master.Free()

	objId := master.Target()
	c, err := repo.LookupCommit(objId)
	if err != nil {
		return err
	}

	parentCommit = append(parentCommit, c)
	_, err = repo.CreateCommit(MasterRef, signature, commitSignature, message, tree, parentCommit...)
	return err
}
