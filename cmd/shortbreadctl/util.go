package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"

<<<<<<< HEAD
	"github.com/coreos/shortbread/api"
=======
	"github.com/coreos/shortbread/client"
>>>>>>> 58afb88... Corrected formatting errors from PR, added revoke and list sub-commands,using the usernames as keys, one command to add and modify an user.
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

<<<<<<< HEAD
func getHTTPClientService() (*api.Service, error) {
=======
func getHTTPClientService() (*client.Service, error) {
>>>>>>> 58afb88... Corrected formatting errors from PR, added revoke and list sub-commands,using the usernames as keys, one command to add and modify an user.
	dialFunc := func(string, string) (net.Conn, error) {
		return net.Dial("tcp", "54.166.129.131:80")
	}

	trans := http.Transport{
		Dial: dialFunc,
	}

	hc := &http.Client{
		Transport: &trans,
	}

<<<<<<< HEAD
	svc, err := api.New(hc)
=======
	svc, err := client.New(hc)
>>>>>>> 58afb88... Corrected formatting errors from PR, added revoke and list sub-commands,using the usernames as keys, one command to add and modify an user.
	if err != nil {
		return nil, err
	}

	(*svc).BasePath = setBasePath()
	return svc, nil
}

func loadPublicKey(path string) string {
<<<<<<< HEAD
=======
	// will catch empty public key error on the server side.
	if path == "" {
		return path
	}

>>>>>>> 58afb88... Corrected formatting errors from PR, added revoke and list sub-commands,using the usernames as keys, one command to add and modify an user.
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
