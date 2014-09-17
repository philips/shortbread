package main

import (
	"fmt"
	"log"
	"os"

	"github.com/coreos/shortbread/Godeps/_workspace/src/github.com/coreos/cobra"
	"github.com/coreos/shortbread/api"
	"github.com/coreos/shortbread/util"
)

var (
	userAdd *cobra.Command
)

func init() {
	userAdd = &cobra.Command{
		Use:   "user-add",
		Short: "associate a user's public key  with a name for easy recall",
		Long:  "associate a user's public key  with a name for easy recall. Example usage: shortbreadctl user-add me path/to/id_rsa.pub",
		Run:   addUserToDirectory,
	}
}

// addUserToDirectory takes in the key value pair provided by the user and adds it to the user directory on the Certifying authority server.
func addUserToDirectory(c *cobra.Command, args []string) {
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "command must have two arguments: The user name and the path to the user's public key.\nType shortbreadctl help user-add for more information")
		return
	}

	key := args[0]
	publicKeyPath := args[1]
	publicKey := util.LoadPublicKey(publicKeyPath)

	svc, err := util.GetHTTPClientService(serverURL)
	if err != nil {
		log.Println(err)
	}

	directoryPair := &api.DirectoryPair{
		Key:          key,
		Value:        publicKey,
		GitSignature: gitSignature,
	}
	err = svc.Directory.UpdateUserDirectory(directoryPair).Do()
	if err != nil {
		log.Println(err)
	}
}
