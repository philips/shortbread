package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"code.google.com/p/go.crypto/ssh"
	"github.com/coreos/shortbread/client"
)

// alias for global map with usernames as keys and slice of certs as value
type CertificateCollection map[string][]*ssh.Certificate

type UserList struct {
	List []string `json:"list,omitempty"`
}

var Certificates CertificateCollection

func init() {
	Certificates = make(CertificateCollection)
}

func (c CertificateCollection) New(certInfo client.CertificateInfo) error {
	privateKeyBytes, err := ioutil.ReadFile(certInfo.PrivateKey)
	if err != nil {
		return err
	}

	authority, err := ssh.ParsePrivateKey(privateKeyBytes) // the private key used to sign the certificate.
	if err != nil {
		return err
	}

	keyToSignBytes := []byte(certInfo.Key)
	keyToSign, _, _, _, err := ssh.ParseAuthorizedKey(keyToSignBytes)
	if err != nil {
		return err
	}

	if keyToSign == nil {
		return errors.New("public key is nil")
	}

	cert := &ssh.Certificate{
		Nonce:       []byte{},
		Key:         keyToSign,
		CertType:    ssh.UserCert,
		KeyId:       "user_" + certInfo.User,
		ValidBefore: ssh.CertTimeInfinity, // this will change in later versions
		Permissions: ssh.Permissions{
			CriticalOptions: map[string]string{},
			Extensions:      map[string]string{},
		},
		ValidPrincipals: []string{certInfo.User},
	}

	for _, perm := range certInfo.Permission.Extensions {
		cert.Permissions.Extensions[perm] = ""
	}

	for _, criticalOpts := range certInfo.Permission.CriticalOptions {
		cert.Permissions.CriticalOptions[criticalOpts] = ""
	}

	err = cert.SignCert(rand.Reader, authority)
	if err != nil {
		return err
	}

	user := certInfo.User
	certs, ok := c[user]
	if !ok {
		c[user] = []*ssh.Certificate{cert}
	} else {
		c[user] = append(certs, cert)
	}

	fp := os.Getenv("HOME") + "/.ssh/id_rsa-cert.pub"
	err = ioutil.WriteFile(fp, ssh.MarshalAuthorizedKey(cert), 0600)
	if err != nil {
		return err
	}

	return nil
}

// Revoke takes an username as argument and deletes all certificates associated with it.
// TODO: add more fine grained deletion allowing them to specify host names.
// eg shortbreadctl revoke -u username123 -h *.example.org will only revoke access to all hosts that match the provided regex.
func (c CertificateCollection) Revoke(revokeInfo client.RevokeCertificate) error {
	user := revokeInfo.User
	_, ok := c[user]
	if !ok {
		return errors.New("username does not exist")
	}

	delete(c, user)
	return nil
}

// SignHandler creates a new certificate from the parameters specified in the request.
func SignHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var params client.CertificateInfo

	err := decoder.Decode(&params)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", err.Error())
		return
	}

	err = Certificates.New(params)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", err.Error())
	}
}

// TODO: abstract the decode code into a common function, will have to use type inference for that to work.
// TODO: verify correct http error code being used.
func RevokeHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var revokeInfo client.RevokeCertificate

	err := decoder.Decode(&revokeInfo)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", err.Error())
		return
	}

	err = Certificates.Revoke(revokeInfo)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", err.Error())
	}
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	fmt.Println("url path is ", url)
	users := new(client.UserList)
	users.List = make([]string, 0)
	for k, _ := range Certificates {
		users.List = append(users.List, k)
	}
	enc := json.NewEncoder(w)
	enc.Encode(users)
}

func main() {
	http.HandleFunc("/v1/sign", SignHandler)
	http.HandleFunc("/v1/revoke", RevokeHandler)
	http.HandleFunc("/v1/get", GetHandler)
	http.ListenAndServe(":8080", nil)
}
