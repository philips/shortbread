package main

import (
	"fmt"
	"os"

	"github.com/coreos/cobra"
	"github.com/coreos/shortbread/client"
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

func issueRevoke(c *cobra.Command, args []string) {
	svc, err := getHTTPClientService() //TODO: modify function to accept a value (user configured base URL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		return
	}

	revokeCrt := &client.RevokeCertificate{
		User: user,
		Key:  loadPublicKey(key),
	}

	crtSvc := client.NewCertService(svc)
	err = crtSvc.Revoke(revokeCrt).Do()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		return
	}
}
