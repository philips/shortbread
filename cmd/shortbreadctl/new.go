package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/coreos/shortbread/Godeps/_workspace/src/code.google.com/p/go.crypto/ssh"

	"github.com/coreos/shortbread/Godeps/_workspace/src/github.com/coreos/cobra"
	"github.com/coreos/shortbread/api"
	"github.com/coreos/shortbread/util"
)

type permissions []string

var (
	newCert         *cobra.Command
	privateKey      string
	validBefore     string // in DD-FullMonth-YYYY format, needs to be converted to unix time to match the specification
	validAfter      string // in DD-FullMonth-YYYY format, needs to be converted to unix time to match the specification
	extensions      permissions
	criticalOptions permissions
	certType        string
	user            string
)

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (i *permissions) String() string {
	return fmt.Sprint(*i)
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (i *permissions) Set(value string) error {
	for _, addr := range strings.Split(value, ",") {
		*i = append(*i, addr)
	}
	return nil
}

func (i *permissions) Type() string {
	return "permissions"
}

func init() {
	newCert = &cobra.Command{
		Use:   "new",
		Short: "generate a new certificate",
		Run:   issueRequest,
	}

	newCert.Flags().StringVarP(&privateKey, "private", "p", "", "specify the name of the private key to be used")
	newCert.Flags().StringVarP(&validBefore, "before", "b", "INFINITY", "specify the date(DD-January-YYYY) upto which the certificate is valid. Default value is  \"INFINITY\" .")
	newCert.Flags().StringVarP(&validAfter, "after", "a", "0", "specify the initial date(DD-January-YYYY) from which the certificate will be valid")
	newCert.Flags().VarP(&extensions, "extensions", "e", "comma separated list of permissions(extesions) to bestow upon the user")
	newCert.Flags().VarP(&criticalOptions, "restrictions", "r", "comma separated list of permissions(restrictions) to place on the user")
	newCert.Flags().StringVarP(&certType, "cert", "c", "USER", "choose from \"USER\" or \"HOST\"")
	newCert.Flags().StringVarP(&user, "username", "u", "", "username of the entity to whom the certificate is issued. Must be a valid username stored in the user directory")

}

func issueRequest(c *cobra.Command, args []string) {
	layout := "2-January-2006"
	svc, err := util.GetHTTPClientService(serverURL)
	if err != nil {
		log.Println(err)
	}

	var validAfterUnixTime uint64 = 0
	var validBeforeUnixTime uint64 = 0

	if validBefore == "INFINITY" {
		validBeforeUnixTime = ssh.CertTimeInfinity
	} else {
		validBeforeUnixTime, err = util.ParseDate(layout, validBefore)
		if err != nil {
			log.Println(err)
		}
	}

	validAfterUnixTime, err = util.ParseDate(layout, validAfter)
	if err != nil {
		log.Println(err)
	}

	crtInfo := &api.CertificateInfoWithGitSignature{
		CertType: certType,
		Permission: &api.Permissions{
			Extensions:      extensions,
			CriticalOptions: criticalOptions,
		},
		User:        user,
		PrivateKey:  privateKey,
		ValidAfter:  validAfterUnixTime,
		ValidBefore: validBeforeUnixTime,

		GitSignature: gitSignature, // see shortbreadctl.go
	}

	crtSvc := api.NewCertService(svc)
	err = crtSvc.Sign(crtInfo).Do()
	if err != nil {
		log.Println(err)
	}
}
