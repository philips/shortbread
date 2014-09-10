package main

import (
	"os"
	"log"

	"github.com/coreos/cobra"
	"github.com/coreos/shortbread/api"
	git "github.com/libgit2/git2go"
)

var (
	shortbreadCtl *cobra.Command
	user          string
	key           string
	baseUrl       string
	config 		 *api.GitSignature
)

func init() {
	shortbreadCtl = &cobra.Command{
		Use:   "shortbreadctl",
		Short: "A command line tool to interact with the CA server and issue/revoke/modify user and host certificates",
	}
	shortbreadCtl.PersistentFlags().StringVarP(&user, "username", "u", "", "username of the entity to whom the certificate is issued")
	shortbreadCtl.PersistentFlags().StringVarP(&key, "key", "k", "", "bears the path to the public key that will be signed by the CA's private key")
	shortbreadCtl.PersistentFlags().StringVarP(&baseUrl, "server", "s", "", "base url for the CA server")
	gitconfig,err := git.NewConfig()
	if err != nil {
		log.Fatalf("unable to create git config object: %s", err.Error())
	}

	gitconfig.AddFile(os.ExpandEnv("$HOME/.gitconfig"), git.ConfigLevelGlobal, false)
	config = gitSignatureFromConfig(gitconfig)
}

func main() {
	shortbreadCtl.AddCommand(updateUser)
	shortbreadCtl.AddCommand(revokeUser)
	shortbreadCtl.Execute()
}

// getSignatureFromConfig returns the api.GitSignature object with a users email and name from the ~/.gitconfig file. Each field is initialized to an empty string by default. 
func gitSignatureFromConfig(config *git.Config) *api.GitSignature{
	name,err := config.LookupString("user.name")
	if err != nil {
		log.Println(err)
	}

	email,err := config.LookupString("user.email")
	if err != nil {
		log.Println(err)
	}

	return &api.GitSignature{
		Name: name, 
		Email: email, 
	}
}
