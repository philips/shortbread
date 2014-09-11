package main

import (
	"fmt"
	"os"
	"strings"

	"code.google.com/p/go.crypto/ssh"

	"github.com/coreos/cobra"
	"github.com/coreos/shortbread/api"
	"github.com/coreos/shortbread/util"
)

type permissions []string

var (
	updateUser      *cobra.Command
	privateKey      string
	validBefore     string // in DD-FullMonth-YYYY format, needs to be converted to unix time to match the specification
	validAfter      string // in DD-FullMonth-YYYY format, needs to be converted to unix time to match the specification
	extensions      permissions
	criticalOptions permissions
	certType        string
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
	updateUser = &cobra.Command{
		Use:   "new",
		Short: "generate a new certificate",
		Run:   issueRequest,
	}

	updateUser.Flags().StringVarP(&privateKey, "private", "p", "", "specify the path of the private key to be used in creating the certificate")
	updateUser.Flags().StringVarP(&validBefore, "before", "b", "0", "specify the date(DD-January-YYYY) upto which the certificate is valid. Specify \"INFINITY\" if you want to issue a certificate that never expires")
	updateUser.Flags().StringVarP(&validAfter, "after", "a", "0", "specify the initial date(DD-January-YYYY) from which the certificate will be valid")
	updateUser.Flags().VarP(&extensions, "extensions", "e", "comma separated list of permissions(extesions) to bestow upon the user")
	updateUser.Flags().VarP(&criticalOptions, "restrictions", "r", "comma separated list of permissions(restrictions) to place on the user")
	updateUser.Flags().StringVarP(&certType, "cert", "c", "USER", "choose from \"USER\" or \"HOST\"")
}

func issueRequest(c *cobra.Command, args []string) {
	layout := "2-January-2006"
	svc, err := util.GetHTTPClientService()
	if err != nil {
		panic(err)
	}

	var validAfterUnixTime uint64 = 0
	var validBeforeUnixTime uint64 = 0

	if validBefore == "INFINITY" {
		validBeforeUnixTime = ssh.CertTimeInfinity
	} else {
		validBeforeUnixTime, err = util.ParseDate(layout, validBefore)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err.Error())
		}
	}

	validAfterUnixTime, err = util.ParseDate(layout, validAfter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
	}

	crtInfo := &api.CertificateInfoWithGitSignature{
		CertType: certType,
		Permission: &api.Permissions{
			Extensions:      extensions,
			CriticalOptions: criticalOptions,
		},
		User:        user,
		Key:         util.LoadPublicKey(key),
		PrivateKey:  privateKey,
		ValidAfter:  validAfterUnixTime,
		ValidBefore: validBeforeUnixTime,

		GitSignature: config, // see shortbreadctl.go
	}

	crtSvc := api.NewCertService(svc)
	err = crtSvc.Sign(crtInfo).Do()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
	}
}
