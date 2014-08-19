package main

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os/exec"
	"strings"
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
		pk := loadPublicKey("/Users/shantanu/.ssh/id_rsa.pub")
		certsWithKey, err := crtSvc.GetCerts(pk).Do()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err.Error())
		}
		// look at the decoded data-structure and write it to disk with name taken from the associated private key
		// TODO make it path agnostic-> more generalc
		prvtKey := certsWithKey.List[0].PrivateKey
		// pKey := strings.SplitN(prvtKey, "/", 5)[4]
		ioutil.WriteFile("/Users/shantanu/.ssh/users_ca-cert.pub", []byte(certsWithKey.List[0].Cert), 0600)
		err = exec.Command("cp", "/Users/shantanu/.ssh/id_rsa", prvtKey).Run()
		if err != nil {
			fmt.Println(err)
		}
		err = exec.Command("ssh-add", prvtKey).Run()
		if err != nil {
			fmt.Println(err)
		}

		break // temp delete this later
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

// Fingerprint for a public key is the md5 sum of the base64 encoded key.
func getFingerPrint(publicKey string) (fp [16]byte, err error) {
	data, err := base64.StdEncoding.DecodeString(strings.Split(publicKey, " ")[1])
	if err != nil {
		return fp, err
	}
	fp = md5.Sum(data)
	return fp, nil
}
