package main

import (
	"log"
	"os"

	"github.com/coreos/shortbread/Godeps/_workspace/src/github.com/coreos/cobra"
	"github.com/coreos/shortbread/api"
	git "github.com/coreos/shortbread/Godeps/_workspace/src/github.com/libgit2/git2go"
)

var (
	shortbreadCtl *cobra.Command
	serverURL     string
	gitSignature  *api.GitSignature
)

const (
	shortbreadctlURL = "SHORTBREADCTL_URL"
)

func init() {
	shortbreadCtl = &cobra.Command{
		Use:   "shortbreadctl",
		Short: "A command line tool to interact with the CA server and issue/revoke/modify user and host certificates",
	}
	gitconfig, err := git.NewConfig()
	if err != nil {
		log.Fatalf("unable to create git gitSignature object: %s", err.Error())
	}

	gitconfig.AddFile(os.ExpandEnv("$HOME/.gitconfig"), git.ConfigLevelGlobal, false)
	gitSignature = gitSignatureFromConfig(gitconfig)

	serverURL = os.Getenv(shortbreadctlURL)
}

func main() {
	shortbreadCtl.AddCommand(newCert)
	shortbreadCtl.AddCommand(revokeCert)
	shortbreadCtl.AddCommand(serverAdd)
	shortbreadCtl.AddCommand(userAdd)
	shortbreadCtl.Execute()
}

// getSignatureFromConfig returns the api.GitSignature object with a users email and name from the ~/.gitconfig file. Each field is initialized to an empty string by default.
func gitSignatureFromConfig(config *git.Config) *api.GitSignature {
	name, err := config.LookupString("user.name")
	if err != nil {
		log.Println(err)
	}

	email, err := config.LookupString("user.email")
	if err != nil {
		log.Println(err)
	}

	return &api.GitSignature{
		Name:  name,
		Email: email,
	}
}
