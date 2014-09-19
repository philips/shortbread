package main

import (
	"log"
	"os"

	"github.com/coreos/shortbread/Godeps/_workspace/src/code.google.com/p/gcfg"

	"github.com/coreos/shortbread/Godeps/_workspace/src/github.com/coreos/cobra"
	"github.com/coreos/shortbread/api"
)

var (
	shortbreadCtl *cobra.Command
	serverURL     string
	gitcfg        *GitConfig
	gitSignature  *api.GitSignature
)

type GitConfig struct {
	User struct {
		Name  string
		Email string
	}
}

const (
	shortbreadctlURL = "SHORTBREADCTL_URL"
)

func init() {
	shortbreadCtl = &cobra.Command{
		Use:   "shortbreadctl",
		Short: "A command line tool to interact with the CA server and issue/revoke/modify user and host certificates",
	}
	gitcfg = new(GitConfig)
	gitSignatureFromConfig(gitcfg, os.ExpandEnv("$HOME/.gitconfig"))
	serverURL = os.Getenv(shortbreadctlURL)
}

func main() {
	shortbreadCtl.AddCommand(newCert)
	shortbreadCtl.AddCommand(serverAdd)
	shortbreadCtl.AddCommand(userAdd)
	shortbreadCtl.Execute()
}

// getSignatureFromConfig returns the *GitConfig object with a users email and name from the ~/.gitconfig file. Each field is initialized to an empty string by default.
func gitSignatureFromConfig(gitcfg *GitConfig, configFile string) {
	err := gcfg.ReadFileInto(gitcfg, configFile)
	if err != nil && gitcfg.User.Email == "" && gitcfg.User.Name == "" {
		log.Printf("Failed to parse gitconfig data: %s\n Using default value for git signature", err)
	}

	gitSignature = &api.GitSignature{
		Name:  gitcfg.User.Name,
		Email: gitcfg.User.Email,
	}
}
