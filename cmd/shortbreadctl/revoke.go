package main

import (
	"fmt"
	"os"

	"github.com/coreos/cobra"
<<<<<<< HEAD
	"github.com/coreos/shortbread/api"
=======
	"github.com/coreos/shortbread/client"
>>>>>>> 58afb88... Corrected formatting errors from PR, added revoke and list sub-commands,using the usernames as keys, one command to add and modify an user.
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
<<<<<<< HEAD
		panic(err)
	}

	revokeCrt := &api.RevokeCertificate{
=======
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		return
	}

	revokeCrt := &client.RevokeCertificate{
>>>>>>> 58afb88... Corrected formatting errors from PR, added revoke and list sub-commands,using the usernames as keys, one command to add and modify an user.
		User: user,
		Key:  loadPublicKey(key),
	}

<<<<<<< HEAD
	crtSvc := api.NewCertService(svc)
	err = crtSvc.Revoke(revokeCrt).Do()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
=======
	crtSvc := client.NewCertService(svc)
	err = crtSvc.Revoke(revokeCrt).Do()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		return
>>>>>>> 58afb88... Corrected formatting errors from PR, added revoke and list sub-commands,using the usernames as keys, one command to add and modify an user.
	}
}
