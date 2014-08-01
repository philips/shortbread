package main

import (
	"code.google.com/p/go.crypto/ssh"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Write a simple http server
// response to the client is only wether or not his request was successfull or what error caused it to fail
// want to maintain state

// first, increment a global variable on every request
// can execute a get to get the variable.

// later: specify which public key to sign, and other parameters, username, permissions, restrictions, etc.
// probably a json object that's marshalled and sent over the wire as an http request.

// server processes that request, creates a certificate, upates the global datastructure

//from the user's computer, periodically execute a get method on your key in the data structure to receive updated copies of stuff.

type CertificateCollection map[string][]*ssh.Certificate

type CertificateParameters struct {
	Username       string
	Permissions    []string // no reason for it to be a map at this stage.
	PrivateKeyPath string
	Key            string // for now it points to the path of the public key to be signed.
}

var Certificates CertificateCollection

func init() {
	Certificates = make(CertificateCollection)
}

func (c CertificateCollection) New(params CertificateParameters) {
	// read private key
	privateKeyBytes, err := ioutil.ReadFile(params.PrivateKeyPath)
	check(err)
	authority, err := ssh.ParsePrivateKey(privateKeyBytes) // the private key used to sign the certificate.
	check(err)
	fmt.Printf("associated public key is: %v ", authority.PublicKey())
	// for now, read in public key to be signed.

	keyToSignBytes, err := ioutil.ReadFile(params.Key)
	check(err)
	keyToSign, comment, _, _, err := ssh.ParseAuthorizedKey(keyToSignBytes)
	check(err)

	if keyToSign == nil {
		panic("public key is nil")
		fmt.Println("comment is ", comment)
	}
	// from the params set the permissions and username
	// valid till infinity for now.

	cert := &ssh.Certificate{
		Nonce:       []byte{},
		Key:         keyToSign, // the public key that will be signed
		CertType:    ssh.UserCert,
		KeyId:       "user_" + params.Username,
		ValidBefore: ssh.CertTimeInfinity,
		Permissions: ssh.Permissions{
			CriticalOptions: map[string]string{},
			Extensions:      map[string]string{},
		},
		ValidPrincipals: []string{params.Username},
	}

	fmt.Println("public key is : ", keyToSign.Type())

	// setting the permissions
	for _, v := range params.Permissions {
		cert.Permissions.Extensions[v] = ""
	}

	err = cert.SignCert(rand.Reader, authority)
	check(err)

	//add newly created cert to the file.

	certs, ok := c[params.Username]

	if !ok {
		// key does not exits
		c[params.Username] = []*ssh.Certificate{cert}
	} else {

		c[params.Username] = append(certs, cert)
	}

	// write signed cert to a file:
	err = ioutil.WriteFile("/Users/shantanu/.ssh/id_rsa-cert-server.pub", ssh.MarshalAuthorizedKey(cert), 0600)

	check(err)

	//once created add it to the map.
	// but for now, also write it to file, so that I can use it to connect to the remote server.

}

func incrementHandler(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	var params CertificateParameters
	err := decoder.Decode(&params)
	check(err)
	Certificates.New(params)
	fmt.Println(Certificates)
	fmt.Fprintf(w, "%d", 200)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	http.HandleFunc("/inc/", incrementHandler)
	http.ListenAndServe(":8080", nil)
}
