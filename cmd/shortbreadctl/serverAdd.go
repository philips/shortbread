package main

import (
	"fmt"
	"log"
	"os"

	"github.com/coreos/cobra"
	"github.com/coreos/shortbread/api"
	"github.com/coreos/shortbread/util"
)

var (
	serverAdd *cobra.Command
)

func init() {
	serverAdd = &cobra.Command{
		Use:   "server-add",
		Short: "associate a servers URL with a name for easy recall. Example usage: shortbreadctl server-add example http://example.com",
		Run:   addServerToDirectory,
	}
}

// addServerToMap takes in the key value pair provided by the user and adds it to the server directory on the Certifying authority server.
func addServerToDirectory(c *cobra.Command, args []string) {
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "command must have two arguments: The server name and the server url.\nType shortbreadctl help server-add for more information")
		return
	}

	key := args[0]
	address := args[1]

	svc, err := util.GetHTTPClientService(serverURL)
	if err != nil {
		log.Println(err)
	}

	directoryPair := &api.DirectoryPair{
		Key:          key,
		Value:        address,
		GitSignature: gitSignature,
	}
	err = svc.Directory.UpdateUserDirectory(directoryPair).Do()
	if err != nil {
		log.Println(err)
	}
}
