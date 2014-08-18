package main

import "os"

// read environment var for location of shortbread server
// read environment variable for lcoation of public key : id_rsa.pub
// use the api to send a get request to the server.
// receive cert wrapped with private key info and flag to tell us if we have to delete or add entries from the ssh-agent
// use private key info to create new copy of id_rsa and write cert
// execute the ssh-add command from within go using the correct file paths.
// remove the tmp  files created, data is in the ssh agent now.

// can store identity indefinitely or for a short period and force the client to pull info again and again.

const SHORTBREADCTL_URL = "SHORTBREADCTL_URL"
const PUBLICKEY_LOCATION = "SHORTBREAD_PUBLIC_KEY"

var serverLocation string
var publicKeyLocation string

func init() {
	serverLocation = os.Getenv(SHORTBREADCTL_URL)
	publicKeyLocation = os.Getenv(PUBLICKEY_LOCATION)
}
