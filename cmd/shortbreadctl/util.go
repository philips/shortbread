package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/coreos/shortbread/client"
)

const SHORTBREADCTL_URL = "SHORTBREADCTL_URL"

type permissions []string

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (i *permissions) String() string {
	return fmt.Sprint(*i)
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (i *permissions) Set(value string) error {
	for _, addr := range strings.Split(value, ",") {
		*i = append(*i, addr)
	}
	return nil
}

func (i *permissions) Type() string {
	return "permissions"
}

func getHTTPClientService() (*client.Service, error) {
	dialFunc := func(string, string) (net.Conn, error) {
		return net.Dial("tcp", "127.0.0.1:8080")
	}

	trans := http.Transport{
		Dial: dialFunc,
	}

	hc := &http.Client{
		Transport: &trans,
	}

	svc, err := client.New(hc)
	if err != nil {
		return nil, err
	}

	(*svc).BasePath = setBasePath()
	return svc, nil
}

func loadPublicKey(path string) string {
	keyToSignBytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return string(keyToSignBytes)
}

// setBasePath sets the BasePath of the API service to the value specified in the
// SHORTBREADCTL_URL environment variable. Default value is "http://localhost:8080/v1/"
func setBasePath() string {
	if basePath := os.Getenv(SHORTBREADCTL_URL); basePath != "" {
		return basePath
	}

	return "http://localhost:8080/v1/"
}
