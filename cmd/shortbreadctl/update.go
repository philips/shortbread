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
	updateUser      *cobra.Command
	privateKey      string
	validBefore     int // in days
	validAfter      int // in days
	extensions      permissions
	criticalOptions permissions
	certType        string
)

func init() {
	updateUser = &cobra.Command{
		Use:   "update",
		Short: "generate a certificate for a new user or modify an existing one",
		Run:   issueRequest,
	}

	updateUser.Flags().StringVarP(&privateKey, "private", "p", "", "specify the path of the private key to be used in creating the certificate")
	updateUser.Flags().IntVarP(&validBefore, "before", "b", 0, "number of days the certificate is valid")
	updateUser.Flags().IntVarP(&validAfter, "after", "a", 0, "number of days before the certificate becomes valid")
	updateUser.Flags().VarP(&extensions, "extensions", "e", "comma separated list of permissions(extesions) to bestow upon the user")
	updateUser.Flags().VarP(&criticalOptions, "restrictions", "r", "comma separated list of permissions(restrictions) to place on the user")
	updateUser.Flags().StringVarP(&certType, "cert", "c", "", "choose from \"USER\" or \"HOST\"")

}

// issueRequest parses the command line flags and issues a request to the server
func issueRequest(c *cobra.Command, args []string) {
	svc, err := getHTTPClientService() //TODO: modify function to accept a value (user configured base URL)
	if err != nil {
		panic(err)
	}

<<<<<<< HEAD
	crtInfo := &api.CertificateInfo{
		CertType: certType, // TODO: warn user about using default value.
		Permission: &api.Permissions{
=======
	crtInfo := &client.CertificateInfo{
		CertType: certType, // TODO: warn user about using default value.
		Permission: &client.Permissions{
>>>>>>> 58afb88... Corrected formatting errors from PR, added revoke and list sub-commands,using the usernames as keys, one command to add and modify an user.
			Extensions:      extensions,
			CriticalOptions: criticalOptions,
		},
		User:       user,
		Key:        loadPublicKey(key),
		PrivateKey: privateKey,
	}

<<<<<<< HEAD
	crtSvc := api.NewCertService(svc)
=======
	crtSvc := client.NewCertService(svc)
>>>>>>> 58afb88... Corrected formatting errors from PR, added revoke and list sub-commands,using the usernames as keys, one command to add and modify an user.
	err = crtSvc.Sign(crtInfo).Do()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
	}
}
