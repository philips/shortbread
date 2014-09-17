package gitutil

import (
	"os"

	"github.com/coreos/shortbread/util"
	git "github.com/coreos/shortbread/Godeps/_workspace/src/github.com/libgit2/git2go"
)

const (
	GitAuthenticationKeyLocation = "SHORTBREAD_GIT"
)

// getCredentials uses the private key location specified by $SHORTBREAD_GIT to create the credentials required for a successful SSH connection.
func getCredentials(url string, username_from_url string, allowed_types git.CredType) (int, *git.Cred) {
	privateKey := util.GetenvWithDefault(GitAuthenticationKeyLocation, os.ExpandEnv("$HOME/.ssh/id_rsa"))
	publicKey := privateKey + ".pub"

	errCode, cred := git.NewCredSshKey(username_from_url, publicKey, privateKey, "")
	if errCode != 0 {
		return errCode, nil
	}

	return errCode, &cred
}
