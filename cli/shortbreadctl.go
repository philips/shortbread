package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/coreos/cobra"
	"github.com/coreos/shortbread/client"
)

type permissions []string

var (
	shortbreadCtl   *cobra.Command
	key             string
	privateKey      string
	validBefore     int // in days
	validAfter      int // in days
	user            string
	extensions      permissions
	criticalOptions permissions
	certType        string
	baseUrl         string // location for the CA server
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

func getHTTPClientService() (*client.Service, error) {
	dialFunc := func(string, string) (net.Conn, error) {
		return net.Dial("tcp", "127.0.0.1:8080")
	}

	trans := http.Transport{
		Dial: dialFunc,
	}

	hc := &http.Client{
		Transport: &trans,
	}

	svc, err := client.New(hc)
	if err != nil {
		return nil, err
	}

	(*svc).BasePath = "http://localhost:8080/v1/"
	return svc, nil
}

func loadPublicKey(path string) string {
	keyToSignBytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return string(keyToSignBytes)
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

func init() {
	shortbreadCtl = &cobra.Command{
		Use:   "shortbreadctl",
		Short: "A command line tool to interact with the CA server and issue/revoke/modify user and host certificates",
		Run:   issueRequest,
	}

	shortbreadCtl.Flags().StringVarP(&key, "key", "k", "", "bears the path to the public key that will be signed by the CA's private key")
	shortbreadCtl.Flags().StringVarP(&privateKey, "private", "p", "", "specify the path of the private key to be used in creating the certificate")
	shortbreadCtl.Flags().IntVarP(&validBefore, "before", "b", 0, "number of days the certificate is valid")
	shortbreadCtl.Flags().IntVarP(&validAfter, "after", "a", 0, "number of days before the certificate becomes valid")
	shortbreadCtl.Flags().StringVarP(&user, "username", "u", "", "username of the entity to whom the certificate is issued")
	shortbreadCtl.Flags().VarP(&extensions, "extensions", "e", "comma separated list of permissions(extesions) to bestow upon the user")
	shortbreadCtl.Flags().VarP(&criticalOptions, "restrictions", "r", "comma separated list of permissions(restrictions) to place on the user")
	shortbreadCtl.Flags().StringVarP(&certType, "cert", "c", "", "choose from \"USER\" or \"HOST\"")
	shortbreadCtl.Flags().StringVarP(&baseUrl, "server", "s", "", "base url for the CA server")
}

func main() {
	shortbreadCtl.Execute()
}

// bin/shortbreadctl -k /Users/shantanu/.ssh/id_rsa.pub -p /Users/shantanu/.ssh/users_ca -u shantanu -e permit-pty -c USER
