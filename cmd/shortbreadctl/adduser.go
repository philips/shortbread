package main

import (
	"github.com/coreos/cobra"
	"github.com/coreos/shortbread/client"
)

var (
	addUser         *cobra.Command
	privateKey      string
	validBefore     int // in days
	validAfter      int // in days
	extensions      permissions
	criticalOptions permissions
	certType        string
	baseUrl         string // location for the CA server

)

func init() {
	addUser = &cobra.Command{
		Use:   "adduser",
		Short: "generate a certificate for a new user",
		Run:   issueRequest,
	}

	addUser.Flags().StringVarP(&privateKey, "private", "p", "", "specify the path of the private key to be used in creating the certificate")
	addUser.Flags().IntVarP(&validBefore, "before", "b", 0, "number of days the certificate is valid")
	addUser.Flags().IntVarP(&validAfter, "after", "a", 0, "number of days before the certificate becomes valid")
	addUser.Flags().VarP(&extensions, "extensions", "e", "comma separated list of permissions(extesions) to bestow upon the user")
	addUser.Flags().VarP(&criticalOptions, "restrictions", "r", "comma separated list of permissions(restrictions) to place on the user")
	addUser.Flags().StringVarP(&certType, "cert", "c", "", "choose from \"USER\" or \"HOST\"")
	addUser.Flags().StringVarP(&baseUrl, "server", "s", "", "base url for the CA server")
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
		panic(err)
	}
}
