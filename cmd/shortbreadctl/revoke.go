package main

import (
	"github.com/coreos/cobra"
)

var (
	revokeUser       *cobra.Command
	userToRevoke     string
	revokedPublicKey string
)

func init() {
	revokeUser = &cobra.Command{
		Use:   "revoke",
		Short: "revoke the certificate issued to a particular user",
		Run:   issueRevoke,
	}
}
