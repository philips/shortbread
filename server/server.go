package main

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"code.google.com/p/go.crypto/ssh"
)

type CertificateCollection map[[16]byte][]*ssh.Certificate

type CertificateParameters struct {
	CertType   string
	User       string
	Permission map[string][]string
	PrivateKey string
	Key        string //public key (base 64 encoded bytes converted to string)
}

type RevokeInfo struct {
	User string
	Key  string
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

	// encoded := strings.Split(params.Key, " ")[1]
	// data, err := base64.StdEncoding.DecodeString(encoded)
	// if err != nil {
	// 	fmt.Println("error:", err)
	// 	return err
	// }
	// fmt.Printf("fingerPrint: % x", md5.Sum(data))

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

	// add newly created cert to the global map (with fingerprint as key) and then write to local disk (for now).
	fingerprint, err := getFingerPrint(params.Key)
	if err != nil {
		return err
	}
	fmt.Printf("Final check: % x", fingerprint)

	certs, ok := c[fingerprint]
	if !ok {
		c[fingerprint] = []*ssh.Certificate{cert}
	} else {
		c[fingerprint] = append(certs, cert)
	}

	err = ioutil.WriteFile(os.ExpandEnv("$HOME/.ssh/id_rsa-cert.pub"), ssh.MarshalAuthorizedKey(cert), 0600)
	if err != nil {
		return err
	}

	return nil
}

// Revoke uses the public key provided in the request to delete the corresponding certificate
// from the map. However, if a username is provided then a certificate is deleted only if the fingerprint is found in the map
// and the user is listed as a valid principal in the certificate. Reports an error otherwise.
func (c CertificateCollection) Revoke(revokeInfo RevokeInfo) error {
	fingerprint, err := getFingerPrint(revokeInfo.Key)
	if err != nil {
		return err
	}

	certs, ok := c[fingerprint]
	if !ok {
		return errors.New("certificate not found, check if you have specified the correct public key")
	}

	user := revokeInfo.User
	checkPrincipal := func(certs []*ssh.Certificate, principal string) bool {
		for _, cert := range certs {
			if cert.ValidPrincipals[0] == principal {
				return true
			}
		}
		return false
	}
	if user != "" && checkPrincipal(certs, user) {
		return errors.New("certificates valid principal differs from username provided")
	}

	delete(c, fingerprint)
	return nil
}

// SignHandler creates a new certificate from the parameters specified in the request.
func SignHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var params CertificateParameters

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
	var revokeInfo RevokeInfo

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

// Fingerprint for a publick key is the md5 sum of the base64 encoded key.
func getFingerPrint(publicKey string) (fp [16]byte, err error) {
	data, err := base64.StdEncoding.DecodeString(strings.Split(publicKey, " ")[1])
	if err != nil {
		return fp, err
	}
	fp = md5.Sum(data)
	return fp, nil
}

func main() {
	http.HandleFunc("/v1/sign", SignHandler)
	http.HandleFunc("/v1/revoke", RevokeHandler)
	http.ListenAndServe(":8080", nil)
}
