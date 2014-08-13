package main

import (
	"fmt"
	"os"

	"github.com/coreos/cobra"
	"github.com/coreos/shortbread/client"
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

	crtInfo := &client.CertificateInfo{
		CertType: certType, // TODO: warn user about using default value.
		Permission: &client.Permissions{
			Extensions:      extensions,
			CriticalOptions: criticalOptions,
		},
		User:       user,
		Key:        loadPublicKey(key),
		PrivateKey: privateKey,
	}

	crtSvc := client.NewCertService(svc)
	err = crtSvc.Sign(crtInfo).Do()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
	}
}
