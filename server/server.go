package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"code.google.com/p/go.crypto/ssh"
)

type CertificateCollection map[string][]*ssh.Certificate

type CertificateParameters struct {
	CertType   string
	User       string
	Permission map[string][]string
	PrivateKey string
	Key        string //public key (base 64 encoded bytes converted to string)
}

var Certificates CertificateCollection

func init() {
	Certificates = make(CertificateCollection)
}

func (c CertificateCollection) New(params CertificateParameters) error {
	privateKeyBytes, err := ioutil.ReadFile(params.PrivateKey)
	if err != nil {
		return err
	}

	authority, err := ssh.ParsePrivateKey(privateKeyBytes) // the private key used to sign the certificate.
	if err != nil {
		return err
	}

	keyToSignBytes := []byte(params.Key)
	keyToSign, _, _, _, err := ssh.ParseAuthorizedKey(keyToSignBytes)
	if err != nil {
		return err
	}

	if keyToSign == nil {
		panic("public key is nil")
	}

	cert := &ssh.Certificate{
		Nonce:       []byte{},
		Key:         keyToSign,
		CertType:    ssh.UserCert,
		KeyId:       "user_" + params.User,
		ValidBefore: ssh.CertTimeInfinity, // this will change in later versions
		Permissions: ssh.Permissions{
			CriticalOptions: map[string]string{},
			Extensions:      map[string]string{},
		},
		ValidPrincipals: []string{params.User},
	}

	for _, perm := range params.Permission["extensions"] {
		cert.Permissions.Extensions[perm] = ""
	}

	for _, criticalOpts := range params.Permission["criticalOptions"] {
		cert.Permissions.CriticalOptions[criticalOpts] = ""
	}

	err = cert.SignCert(rand.Reader, authority)
	if err != nil {
		return err
	}

	//add newly created cert to the global map + write to local disk (for now).
	certs, ok := c[params.User]
	if !ok {
		c[params.User] = []*ssh.Certificate{cert}
	} else {
		c[params.User] = append(certs, cert)
	}

	err = ioutil.WriteFile("/Users/shantanu/.ssh/id_rsa-cert-server1.pub", ssh.MarshalAuthorizedKey(cert), 0600)
	if err != nil {
		return err
	}

	return nil
}

// SignHandler creates a new certificate from the parameters specified in the request.
func SignHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var params CertificateParameters

	err := decoder.Decode(&params)
	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
		return
	}

	err = Certificates.New(params)
	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
	}
}

func main() {
	http.HandleFunc("/v1/", SignHandler)
	http.ListenAndServe(":8080", nil)
}
