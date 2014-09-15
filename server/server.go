package main

import (
	"crypto/rand"
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

type CertificatesAndMetaData struct {
	cert       *ssh.Certificate
	privateKey string
}
type Fingerprint [16]byte
type CertificateCollection map[Fingerprint][]*CertificatesAndMetaData
type Directory map[string]string

const (
	serverDirectoryFile = "serverDirectory"
	userDirectoryFile   = "userDirectory"
)

var (
	Certificates    CertificateCollection
	remoteRepoUrl   string
	path            string = filepath.Join(os.Getenv("HOME"), "ssh", "shortbread/certs/.git")
	gitWorkingDir   string
	mutex           = &sync.Mutex{}
	serverDirectory Directory
	userDirectory   Directory
)

func init() {
	Certificates = make(CertificateCollection)
	if len(os.Args) >= 2 {
		remoteRepoUrl = os.Args[1]
		log.Printf("remote repo specified to be: %s\n", remoteRepoUrl)
	}
	serverDirectory = make(Directory)
	userDirectory = make(Directory)
	Certificates.initialize()
}

// Initalize the map based on existing contents of the local git repo, if one exists.
func (c CertificateCollection) initialize() {
	repo, err := gitutil.OpenRepository(remoteRepoUrl, path)
	if err != nil {
		return
	}
	defer repo.Free()
	gitWorkingDir = repo.Workdir()

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
		if certPath == serverDirectoryFile {
			err := readServerDirectory(&serverDirectory, gitWorkingDir)
			if err != nil {
				log.Println("Server directory is empty ! Could not load directory from disk: ", err)
			}

		}

		if certPath == userDirectoryFile {
			err := readUserDirectory(&userDirectory, gitWorkingDir)
			if err != nil {
				log.Println("User directory is empty ! Could not load directory from disk: ", err)
			}
		}
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

	repo, err := gitutil.OpenRepository(remoteRepoUrl, path)
	if err != nil {
		log.Print(err)
		return err
	}
	defer repo.Free()

	privateKeyBytes, err := ioutil.ReadFile(filepath.Join(os.ExpandEnv("$HOME/ssh"), params.PrivateKey))
	if err != nil {
		return err
	}

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

	if remoteRepoUrl != "" {
		err = gitutil.Push(repo)
		if err != nil {
			log.Printf("Push to remote repo failed: %s\n", err.Error())
		}
		log.Printf("Pushed to %s\n", remoteRepoUrl)
	}

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

func ServerDirectoryHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var dirPair api.DirectoryPair

	err := decoder.Decode(&dirPair)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", err.Error())
		return
	}
	key := dirPair.Key
	address := dirPair.Value

	if val, ok := serverDirectory[key]; ok {
		w.WriteHeader(http.StatusConflict)
		fmt.Fprintf(w, "key %s already exists and is mapped to:  %s", key, val)
		return
	}

	// update directory and commit to git
	serverDirectory[key] = address

	mutex.Lock()
	defer mutex.Unlock()

	err = writeServerDirectory(&serverDirectory, gitWorkingDir)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Could not write directory to disk: %s", err.Error())
		return
	}

	authorName := dirPair.GitSignature.Name
	authorEmail := dirPair.GitSignature.Email
	err = addAndCommitDirectory(serverDirectoryFile, authorName, authorEmail, fmt.Sprintf("Added new entry to server directory: %s = %s ", key, address))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Could not commit directory into git: %s", err.Error())
		return
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

func main() {
	http.HandleFunc("/v1/sign", SignHandler)
	http.HandleFunc("/v1/revoke", RevokeHandler)
	http.HandleFunc("/v1/getcerts/", ClientHandler)
	http.HandleFunc("/v1/updateServerDirectory", ServerDirectoryHandler)
	http.ListenAndServe(":8080", nil)
}
