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

	"code.google.com/p/go.crypto/ssh"
	"github.com/coreos/shortbread/client"
)

type CertificateCollection map[[16]byte][]*CertificatesAndMetaData

var Certificates CertificateCollection

func init() {
	Certificates = make(CertificateCollection)
}

func (c CertificateCollection) New(params api.CertificateInfo) error {
	privateKeyBytes, err := ioutil.ReadFile(os.ExpandEnv("$SHORTBREAD_PRVT_KEY") + string(os.PathSeparator) + params.PrivateKey)
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

	certAndKey := &CertificatesAndMetaData{
		signedCert: cert,
		changed:    true,
		transferData: &api.CertificateAndPrivateKey{
			Cert:       string(ssh.MarshalAuthorizedKey(cert)),
			PrivateKey: params.PrivateKey,
		},
	}
	// add newly created cert to the global map (with fingerprint as key) and then write to local disk (for now).
	fingerprint, err := getFingerPrint(params.Key)
	if err != nil {
		return err
	}

	certs, ok := c[fingerprint]
	if !ok {
		c[fingerprint] = []*CertificatesAndMetaData{certAndKey}
	} else {
		c[fingerprint] = append(certs, certAndKey)
	}

	// err = ioutil.WriteFile(os.ExpandEnv("$HOME/.ssh/id_rsa-cert.pub"), ssh.MarshalAuthorizedKey(cert), 0600)
	// if err != nil {
	// 	return err
	// }

	return nil
}

// Revoke uses the public key provided in the request to delete the corresponding certificate
// from the map. However, if a username is provided then a certificate is deleted only if the fingerprint is found in the map
// and the user is listed as a valid principal in the certificate. Reports an error otherwise.
func (c CertificateCollection) Revoke(revokeInfo api.RevokeCertificate) error {
	fingerprint, err := getFingerPrint(revokeInfo.Key)
	if err != nil {
		return err
	}

	certs, ok := c[fingerprint]
	if !ok {
		return errors.New("certificate not found, check if you have specified the correct public key")
	}

	user := revokeInfo.User
	checkPrincipal := func(certs []*CertificatesAndMetaData, principal string) bool {
		for _, certKey := range certs {
			if certKey.signedCert.ValidPrincipals[0] == principal {
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
	var params api.CertificateInfo

	err := decoder.Decode(&params)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", err.Error())
		return
	}

	err = Certificates.New(params)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
<<<<<<< HEAD
		fmt.Fprintf(w, "%s", err.Error())
	}
}

// TODO: abstract the decode code into a common function, will have to use type inference for that to work.
// TODO: verify correct http error code being used.
func RevokeHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var revokeInfo api.RevokeCertificate

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

func ClientHandler(w http.ResponseWriter, r *http.Request) {
	fingerprint, _ := getFingerPrint(strings.SplitN(r.URL.Path, "/", 4)[3])
	certs, ok := Certificates[fingerprint]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", "key does not exist.")
		return
	}

	// encode the list here
	certsWithKey := new(api.CertificatesWithKey)
	// certsWithKey.List = new([]api.CertificateAndPrivateKey)
	for _, t := range certs {
		certsWithKey.List = append(certsWithKey.List, t.transferData)
	}
	enc := json.NewEncoder(w)
	enc.Encode(certsWithKey)

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
	http.HandleFunc("/v1/getcerts/", ClientHandler)
	http.ListenAndServe(":8080", nil)
}
