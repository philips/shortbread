package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/coreos/cobra"
)

type directory map[string]string

var (
	serverAdd       *cobra.Command
	serverDirectory *directory
)

const (
	serverDirectoryFile = os.Getenv("$HOME/serverMapFile")
)

func init() {
	serverAdd = &cobra.Command{
		Use:   "server-add",
		Short: "associate a servers URL with a name for easy recall. Example usage: shortbreadctl server-add me http://127.0.0.1:8888/api/",
		Run:   addServerToDirectory,
	}
	serverDirectory = new(serverDirectory)
	serverDirectory.initDirectory(serverMapFile)
}

// addServerToMap takes in the key value pair provided by the user and adds it to the serverDirectory.
// The data is also backed up on disk.
func addServerToDirectory(c *cobra.Command, args []string) {
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "command must have two arguments: type shortbreadctl help server-add for more information")
		return
	}

	key := args[0]
	address := args[1]

	if server, ok := serverDirectory[key]; ok {
		oldAddress := server
		fmt.Printf("warning: overriding existing key-value pair %s:%s with %s:%s", key, oldAddress, key, address)
	}

	serverDirectory[key] = address
	serverDirectory.writeDirectoryToDisk(serverDirectoryFile)
}

// initMap parses the encoded content in filePath and uses it to initialize the serverDirectory.
// If the file does not exist, then it returns an empty map.
func (directory *directory) initDirectory(filePath) {
	encodedMap, err := ioutil.ReadFile(filePath)

	_, ok := err.(*os.PathError)
	if err != nil && ok {
		log.Println(err.Error())
		return
	}

	if err != nil {
		log.Fatalln(err.Error())
	}

	encdedMapReader = bytes.NewReader(encodedMap)
	dec := gob.NewDecoder(encodedMapReader)
	err = dec.Decode(directory)
	if err != nil {
		log.Fatal("decode error: ", err)
	}
}

// writeDirectoryToDisk encodes the contents of the directory map and writes to the file specified by filePath.
func (directory *directory) writeDirectoryToDisk(filePath) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(directory)
	if err != nil {
		log.Fatal("failed to encode map: ", err)
	}

	err = ioutil.WriteFile(filePath, buffer.Bytes(), 0644)
	if err != nil {
		log.Fatal("failed to write encoded map to disk: ", err)
	}
}
