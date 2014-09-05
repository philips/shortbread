package gitutil

import (
	"log"
	"strings"

	git "github.com/libgit2/git2go"
)

// OpenRepository returns a pointer to the local repo specified by `path`. If a local repo does not exist then it creates one by cloning the repo located at `url`
func OpenRepository(url string, path string) (*git.Repository, error) {
	repo, err := git.OpenRepository(path)
	if err != nil {
		log.Print("repo does not exist: creating one")
		repo, err = gitClone(url, strings.Split(path, "/.git")[0])
		if err != nil {
			return nil, err
		}
	}
	return repo, err
}

func gitClone(url string, path string) (*git.Repository, error) {
	options := &git.CloneOptions{
		Bare:             false,
		IgnoreCertErrors: false,
		RemoteName:       "origin",
		CheckoutBranch:   "master",
		RemoteCallbacks: &git.RemoteCallbacks{
			CredentialsCallback: getCredentials,
		},
		CheckoutOpts: &git.CheckoutOpts{
			Strategy: git.CheckoutSafeCreate,
		},
	}
	log.Printf("%s\n%s", url, path)
	repo, err := git.Clone(url, path, options)
	if err != nil {
		return nil, err
	}
	return repo, nil
}
