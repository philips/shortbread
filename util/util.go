package util

import (
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"code.google.com/p/go.crypto/ssh"

	"github.com/coreos/shortbread/api"
)

const (
	defaultURL = "http://localhost:8889/v1/"
)

func GetHTTPClientService(serverURL string) (*api.Service, error) {
	if serverURL == "" {
		serverURL = defaultURL
	}
	return getHTTPClientService(serverURL)
}

func getHTTPClientService(basePath string) (*api.Service, error) {
	dialFunc := func(string, string) (net.Conn, error) {
		addr, err := setAddress(basePath)
		if err != nil {
			return nil, err
		}
		return net.Dial("tcp", addr)
	}

	trans := http.Transport{
		Dial: dialFunc,
	}

	hc := &http.Client{
		Transport: &trans,
	}

	svc, err := api.New(hc)
	if err != nil {
		return nil, err
	}

	(*svc).BasePath = basePath
	return svc, nil
}

func LoadPublicKey(path string) string {
	keyToSignBytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return string(keyToSignBytes)
}

func ParseSSHCert(rawCert []byte) (*ssh.Certificate, error) {
	certPubKey, _, _, _, err := ssh.ParseAuthorizedKey(rawCert)
	if err != nil {
		return nil, err
	}
	cert := certPubKey.(*ssh.Certificate)
	return cert, nil
}

// setAddress accepts the basepath as input and extracts the hostname and port number from the url.
func setAddress(basePath string) (string, error) {
	addr, err := url.Parse(basePath)
	if err != nil {
		return "", err
	}
	return addr.Host, nil
}

// GetenvWithDefault reads in the value of an environment variable and if it is undefined retuns the default value.
func GetenvWithDefault(variable, defaultValue string) string {
	v := os.Getenv(variable)
	if v != "" {
		return v
	}
	return defaultValue
}

// ParseDate converts the date into Unix Time (time since 1st Jan 1970 in seconds)
func ParseDate(layout, value string) (uint64, error) {
	if value == "0" {
		return 0, nil
	}

	t, err := time.Parse(layout, value)
	if err != nil {
		return 0, err
	}

	return uint64(t.Unix()), nil
}
