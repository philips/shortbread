package main

import (
	"log"

	"github.com/coreos/shortbread/Godeps/_workspace/src/github.com/coreos/cobra"
)

var (
	revokeCert       *cobra.Command
	userToRevoke     string
	revokedPublicKey string
)

func init() {
	revokeCert = &cobra.Command{
		Use:   "revoke",
		Short: "revoke the certificate issued to a particular user",
		Run:   issueRevoke,
	}
}

func issueRevoke(c *cobra.Command, args []string) {
	log.Println("not implemented yet.")
}
