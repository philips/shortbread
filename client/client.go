package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/go.crypto/ssh/agent"

	"github.com/coreos/shortbread/api"
	"github.com/coreos/shortbread/util"
)

const (
	SHORTBREAD_PUBLIC_KEY = "SHORTBREAD_PUBLIC_KEY"
)

func main() {
	svc, err := util.GetHTTPClientService()
	if err != nil {
		log.Println(err)
	}

	crtSvc := api.NewCertService(svc)
	publicKeyPath := util.GetenvWithDefault(SHORTBREAD_PUBLIC_KEY, os.ExpandEnv("$HOME/.ssh/id_rsa.pub"))
	privateKeyPath := strings.Split(publicKeyPath, ".pub")[0]
	pk := util.LoadPublicKey(publicKeyPath)

	for {
		time.Sleep(2000 * time.Millisecond)
		certsWithKey, err := crtSvc.GetCerts(pk).Do()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err.Error())
		}
		err = updateSSHAgent(certsWithKey.List, privateKeyPath)
		if err != nil {
			log.Println(err)
		}
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
		err = sshAgent.Add(privateKeyInterface, cert, "certificated added by shortbread")
		if err != nil {
			return err
		}
	}
	return nil
}
