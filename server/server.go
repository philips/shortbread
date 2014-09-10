package main

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"code.google.com/p/go.crypto/ssh"
	"github.com/coreos/shortbread/api"
	"github.com/coreos/shortbread/gitutil"
	"github.com/coreos/shortbread/util"
)

const (
	SHORTBREAD_PRVT_KEY = "SHORTBREAD_PRVT_KEY"
	FingerprintLength   = 16
)

type CertificatesAndMetaData struct {
	cert       *ssh.Certificate
	privateKey string
}
type Fingerprint [16]byte 
type CertificateCollection map[Fingerprint][]*CertificatesAndMetaData

var Certificates CertificateCollection
var url string = "git@github.com:joshi4/shortbread-test.git"
var path string = filepath.Join(os.Getenv("HOME"), "ssh","shortbread/certs/.git") 
var mutex = &sync.Mutex{}

func init() {
	Certificates = make(CertificateCollection)
	Certificates.initialize()
}

// Initalize the map based on existing contents of the local git repo, if one exists.
func (c CertificateCollection) initialize() {
	repo, err := gitutil.OpenRepository(url, path)
	if err != nil {
		return
	}
	defer repo.Free()

	index, err := repo.Index()
	if err != nil {
		return
	}
	defer index.Free()

	entryCount := index.EntryCount()
	var i uint
	for i = 0; i < entryCount; i++ {
		indexEntry, err := index.EntryByIndex(i)
		if err != nil {
			continue
		}
		//add to the map only if path ends with `.pub` and after splitting on pathSeparator, length is 2.
		certPath := indexEntry.Path
		certPathSlice := strings.Split(certPath, string(os.PathSeparator))
		dirName := certPathSlice[0]
		if strings.HasSuffix(certPath, "-cert.pub") && len(certPathSlice) == 2 && len(dirName) == 32 {
			certName := certPathSlice[1]
			var fingerprint Fingerprint
			hexadecimal := fingerprint[:]
			hexadecimal, err := hex.DecodeString(dirName)
			if err != nil {
				continue
			}
			copy(fingerprint[:], hexadecimal)
			rawCert, err := ioutil.ReadFile(filepath.Join(repo.Workdir(), certPath))
			if err != nil {
				continue
			}

			cert, err := util.ParseSSHCert(rawCert)
			if err != nil {
				continue
			}

			certs, ok := c[fingerprint]
			certAndKey := &CertificatesAndMetaData{
				cert:       cert,
				privateKey: strings.Split(certName, "-cert.pub")[0],
			}
			if !ok {
				c[fingerprint] = []*CertificatesAndMetaData{certAndKey}
			} else {
				c[fingerprint] = append(certs, certAndKey)
			}

		}
	}
}

// New creates a new certificate based on the information supplied by the user and adds it to the global map.
// Each new entry is logged in a git repo. 
func (c CertificateCollection) New(params api.CertificateInfoWithGitSignature) error {
	mutex.Lock()
	defer mutex.Unlock()

	repo, err := gitutil.OpenRepository(url, path)
	if err != nil {
		log.Print(err)
		return err
	}
	defer repo.Free()

	privateKeyBytes, err := ioutil.ReadFile(filepath.Join(util.GetenvWithDefault(SHORTBREAD_PRVT_KEY, os.ExpandEnv("$HOME/ssh")), params.PrivateKey))
	if err != nil {
		return err
	}

	//the private key used to sign the certificate.
	authority, err := ssh.ParsePrivateKey(privateKeyBytes)
	if err != nil {
		return err
	}

	keyToSignBytes := []byte(params.Key)
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
		KeyId:       "user_" + params.User,
		ValidBefore: params.ValidBefore,
		ValidAfter:  params.ValidAfter,
		Permissions: ssh.Permissions{
			CriticalOptions: map[string]string{},
			Extensions:      map[string]string{},
		},
		ValidPrincipals: []string{params.User},
	}

	for _, perm := range params.Permission.Extensions {
		cert.Permissions.Extensions[perm] = ""
	}

	for _, criticalOpts := range params.Permission.CriticalOptions {
		cert.Permissions.CriticalOptions[criticalOpts] = ""
	}

	err = cert.SignCert(rand.Reader, authority)
	if err != nil {
		return err
	}

	certAndKey := &CertificatesAndMetaData{
		cert:       cert,
		privateKey: params.PrivateKey,
	}

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
	certDirPath := filepath.Join(repo.Workdir(), fmt.Sprintf("%x", fingerprint))
	fileInfo, err := os.Stat(certDirPath)
	if err != nil || !fileInfo.IsDir() {
		err := os.Mkdir(certDirPath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	certPath := filepath.Join(certDirPath, (params.PrivateKey + "-cert.pub"))
	err = ioutil.WriteFile(certPath, ssh.MarshalAuthorizedKey(cert), 0600)
	if err != nil {
		return err
	}

	relativeCertPath := filepath.Join(fmt.Sprintf("%x", fingerprint), (params.PrivateKey + "-cert.pub"))

	err = gitutil.AddAndCommit(repo, []string{relativeCertPath}, fmt.Sprintf("added cert for user: %s with private key name: %s", params.User, params.PrivateKey), params.GitSignature.Name, params.GitSignature.Email)
	if err != nil {
		return err
	}

	// err = gitutil.Push(repo)
	// if err != nil {
	// 	return err
	// }

	return nil
}

// Revoke uses the public key provided in the request to delete the corresponding certificate
// from the map. If an username is provided then a certificate is deleted only if it's listed as a valid principal
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
		for _, certData := range certs {
			if certData.cert.ValidPrincipals[0] == principal {
				return true
			}
		}
		return false
	}
	if user != "" && checkPrincipal(certs, user) {
		return errors.New("certificate's valid principal differs from username provided")
	}

	delete(c, fingerprint)
	return nil
}

// SignHandler creates a new certificate from the parameters specified in the request.
func SignHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var params api.CertificateInfoWithGitSignature

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

	certsWithKey := new(api.CertificatesWithKey)
	for _, certAndKey := range certs {
		transferData := &api.CertificateAndPrivateKey{
			Cert:       string(ssh.MarshalAuthorizedKey(certAndKey.cert)),
			PrivateKey: certAndKey.privateKey,
		}
		certsWithKey.List = append(certsWithKey.List, transferData)
	}
	enc := json.NewEncoder(w)
	enc.Encode(certsWithKey)

}

// Fingerprint for a public key is the md5 sum of the base64 encoded key.
func getFingerPrint(publicKey string) (fp Fingerprint, err error) {
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
	http.HandleFunc("/v1/getcerts/", ClientHandler)
	http.ListenAndServe(":8080", nil)
}