package main

import (
	"fmt"
	"os"

	"github.com/coreos/cobra"
	"github.com/coreos/shortbread/api"
	"github.com/coreos/shortbread/util"
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
	svc, err := util.GetHTTPClientService()
	if err != nil {
		panic(err)
	}

	revokeCrt := &api.RevokeCertificate{
		User: user,
		Key:  util.LoadPublicKey(key),
	}

	crtSvc := api.NewCertService(svc)
	err = crtSvc.Revoke(revokeCrt).Do()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
	}
}
