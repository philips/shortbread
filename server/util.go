package main

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/gob"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/coreos/shortbread/gitutil"
)

// Fingerprint for a public key is the md5 sum of the base64 encoded key.
func getFingerPrint(publicKey string) (fp Fingerprint, err error) {
	data, err := base64.StdEncoding.DecodeString(strings.Split(publicKey, " ")[1])
	if err != nil {
		return fp, err
	}
	fp = md5.Sum(data)
	return fp, nil
}

// writeDirectory takes the full path to the file and encoded the directory prior to writing it to disk.
func (directory *Directory) writeDirectory(file string) error {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(directory)
	if err != nil {
		log.Println("Could not encode repo")
		return err
	}

	err = ioutil.WriteFile(file, buffer.Bytes(), 0644)
	if err != nil {
		log.Println("failed to write encoded map to disk")
		return err
	}
	return nil
}

func writeServerDirectory(directory *Directory, workingDir string) error {
	err := directory.writeDirectory(filepath.Join(workingDir, serverDirectoryFile))
	return err
}

func writeUserDirectory(directory *Directory, workingDir string) error {
	err := directory.writeDirectory(filepath.Join(workingDir, userDirectoryFile))
	return err
}

func (directory *Directory) readDirectory(file string) error {
	encodedMap, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	encodedMapReader := bytes.NewReader(encodedMap)
	dec := gob.NewDecoder(encodedMapReader)
	err = dec.Decode(directory)

	return err
}

func readServerDirectory(directory *Directory, workingDir string) error {
	return directory.readDirectory(filepath.Join(workingDir, serverDirectoryFile))
}

func readUserDirectory(directory *Directory, workingDir string) error {
	return directory.readDirectory(filepath.Join(workingDir, userDirectoryFile))
}

func addAndCommitDirectory(fileName, authorName, authorEmail, msg string) error {
	repo, err := gitutil.OpenRepository(remoteRepoUrl, path)
	if err != nil {
		return err
	}

	err = gitutil.AddAndCommit(repo, []string{fileName}, msg, authorName, authorEmail)
	return err
}
