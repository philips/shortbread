package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/coreos/shortbread/api"

	"os"
)

// read environment var for location of shortbread server
// read environment variable for lcoation of public key : id_rsa.pub
// use the api to send a get request to the server.
// receive cert wrapped with private key info and flag to tell us if we have to delete or add entries from the ssh-agent
// use private key info to create new copy of id_rsa and write cert
// execute the ssh-add command from within go using the correct file paths.
// remove the tmp  files created, data is in the ssh agent now.

// can store identity indefinitely or for a short period and force the client to pull info again and again.

const SHORTBREADCTL_URL = "SHORTBREADCTL_URL"
const PUBLICKEY_LOCATION = "SHORTBREAD_PUBLIC_KEY"

var serverLocation string
var publicKeyLocation string

func init() {
	serverLocation = os.Getenv(SHORTBREADCTL_URL)
	publicKeyLocation = os.Getenv(PUBLICKEY_LOCATION)
}

func main() {
	// var certWithKey *api.CertificatesWithKey

	svc, err := getHTTPClientService() //TODO: modify function to accept a value (user configured base URL)
	if err != nil {
		panic(err)
	}
	crtSvc := api.NewCertService(svc)

	for {
		time.Sleep(1000 * time.Millisecond)
		_, err := crtSvc.GetCerts((loadPublicKey("/Users/shantanu/.ssh/id_rsa.pub"))).Do()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err.Error())
		}

	}
}

func getHTTPClientService() (*api.Service, error) {
	dialFunc := func(string, string) (net.Conn, error) {
		return net.Dial("tcp", "127.0.0.1:8080")
	}

	trans := http.Transport{
		Dial: dialFunc,
	}

	hc := &http.Client{
		Transport: &trans,
	}

	svc, err := api.New(hc)
	if err != nil {
		return nil, err
	}

	(*svc).BasePath = setBasePath()
	return svc, nil
}

func loadPublicKey(path string) string {
	keyToSignBytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return string(keyToSignBytes)
}

// setBasePath sets the BasePath of the API service to the value specified in the
// SHORTBREADCTL_URL environment variable. Default value is "http://localhost:8080/v1/"
func setBasePath() string {
	if basePath := os.Getenv(SHORTBREADCTL_URL); basePath != "" {
		return basePath
	}

	return "http://localhost:8080/v1/"
}
