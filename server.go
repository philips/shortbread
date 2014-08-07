package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"code.google.com/p/go.crypto/ssh"
)

// Big picture:

// Write a simple http server
// response to the client is only wether or not his request was successfull or what error caused it to fail
// want to maintain state

//specify which public key to sign, and other parameters, username, permissions, restrictions, etc.
// probably a json object that's marshalled and sent over the wire as an http request.

// server processes that request, creates a certificate, upates the global datastructure

//from the user's computer, periodically execute a get method on your key in the data structure to receive updated copies of stuff.

type CertificateCollection map[string][]*ssh.Certificate

type CertificateParameters struct {
	CertType   string
	User       string
	Permission map[string][]string // no reason for it to be a map at this stage.
	PrivateKey string
	Key        string // for now it points to the path of the public key to be signed.
}

var Certificates CertificateCollection

func init() {
	Certificates = make(CertificateCollection)
}

func (c CertificateCollection) New(params CertificateParameters) {
	// read private key
	privateKeyBytes, err := ioutil.ReadFile(params.PrivateKey)
	check(err)
	authority, err := ssh.ParsePrivateKey(privateKeyBytes) // the private key used to sign the certificate.
	check(err)
	fmt.Printf("associated public key is: %v ", authority.PublicKey())
	// for now, read in public key to be signed.

	keyToSignBytes := []byte(params.Key)
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
		KeyId:       "user_" + params.User,
		ValidBefore: ssh.CertTimeInfinity,
		Permissions: ssh.Permissions{
			CriticalOptions: map[string]string{},
			Extensions:      map[string]string{},
		},
		ValidPrincipals: []string{params.User},
	}

	fmt.Println("public key is : ", keyToSign.Type())

	// setting the permissions; // CHANGE THIS
	for _, v := range params.Permission {
		for _, perm := range v {
			cert.Permissions.Extensions[perm] = ""
		}
	}

	err = cert.SignCert(rand.Reader, authority)
	check(err)

	//add newly created cert to the file.

	certs, ok := c[params.User]

	if !ok {
		// key does not exits
		c[params.User] = []*ssh.Certificate{cert}
	} else {

		c[params.User] = append(certs, cert)
	}

	// write signed cert to a file:
	err = ioutil.WriteFile("/Users/shantanu/.ssh/id_rsa-cert-server1.pub", ssh.MarshalAuthorizedKey(cert), 0600)

	check(err)

	//once created add it to the map.
	// but for now, also write it to file, so that I can use it to connect to the remote server.

}

func SignHandler(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	fmt.Println(r.Body)
	var params CertificateParameters
	err := decoder.Decode(&params)
	check(err)
	fmt.Printf("%v", params)
	Certificates.New(params)
	fmt.Printf("%v", Certificates)
	fmt.Fprintf(w, "%d", 200)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	http.HandleFunc("/", SignHandler)
	http.ListenAndServe(":8080", nil)
}

// curl -v -H "Accept: application/json" -H "Content-type: application/json" -X POST -d ' {"User": "shantanu", "Permissions": ["permit-pty"], "PrivateKeyPath": "/Users/shantanu/.ssh/users_ca","Key": "/Users/shantanu/.ssh/id_rsa.pub" } '  http://localhost:8080/sign/
