package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"

	"github.com/coreos/shortbread/Godeps/_workspace/src/code.google.com/p/go.crypto/ssh"
	"github.com/coreos/shortbread/Godeps/_workspace/src/code.google.com/p/go.crypto/ssh/agent"

	"github.com/coreos/shortbread/api"
	"github.com/coreos/shortbread/util"
)

func main() {
	if len(os.Args) < 2 {
		log.Println("Usage: client http://server:port/v1/")
		os.Exit(2)
	}
	serverURL := os.Args[1]

	svc, err := util.GetHTTPClientService(serverURL)
	if err != nil {
		log.Printf("call to util.GetHTTPClientService failed: %s\n", err.Error())
		return
	}

	crtSvc := api.NewCertService(svc)
	// TODO allow user to specify multiple keys instead of enforcing id_rsa.pub
	publicKeyPath := os.ExpandEnv("$HOME/.ssh/id_rsa.pub")
	privateKeyPath := strings.Split(publicKeyPath, ".pub")[0]
	pk := util.LoadPublicKey(publicKeyPath)

	certsWithKey, err := crtSvc.GetCerts(pk).Do()
	if err != nil {
		log.Printf("Get request to API failed: %s\n", err.Error())
		return
	}
	err = updateSSHAgent(certsWithKey.List, privateKeyPath)
	if err != nil {
		log.Printf("Failed to updateSSHAgent: %s\n", err.Error())
	}
}

// updateSSHAgent takes the list of certificates and path to the private key (corresponding to the signed public key). Adds the cert if it's not present in the agent.
func updateSSHAgent(certsWithKeyList []*api.CertificateAndPrivateKey, privateKeyPath string) error {
	conn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return err
	}
	defer conn.Close()

	sshAgent := agent.NewClient(conn)
	certsInSSHAgent, err := sshAgent.List()
	if err != nil {
		return err
	}

	privateKeyBytes, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return err
	}

	privateKeyInterface, err := ssh.ParseRawPrivateKey(privateKeyBytes)
	if err != nil {
		return err
	}

	// TODO optimize this. currently O(N^2)
	for _, certAndKey := range certsWithKeyList {
		cert, err := util.ParseSSHCert([]byte(certAndKey.Cert))
		for _, key := range certsInSSHAgent {
			certBlob := key.Blob
			if err != nil {
				return err
			}
			if bytes.Equal(certBlob, cert.Marshal()) {
				break
			}
		}
		err = sshAgent.Add(privateKeyInterface, cert, "certificate added by shortbread")
		if err != nil {
			return err
		}
	}
	return nil
}
