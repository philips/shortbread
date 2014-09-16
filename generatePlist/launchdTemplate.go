package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	errorLog     = "shortbread.client_error.log"
	outputLog    = "shortbread.client_output.log"
	templateFile = "com.shortbread.plist.template"
	plistFile    = "com.shortbread.client.plist"
)

type PlistTemplate struct {
	OutputFileName      string
	PathToClientBinary  string
	ShortbreadServerURL string
	PathToErrorLog      string
	PathToOutputLog     string
	TimeInterval        string
}

func main() {

	if len(os.Args) != 2 {
		log.Fatalln("Must provide url of shortbread CA: Example usage: generatePlist http://example.com:8889/v1/")
	}

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalln("could not get PWD: ", err)
	}
	templatePathRoot := strings.Split(pwd, "bin")
	basePath := templatePathRoot[0]
	filePathTemplate := filepath.Join(basePath, "script", templateFile)
	filePathPlist := filepath.Join(basePath, "script", plistFile)
	filePathErrorLog := filepath.Join(basePath, "script", errorLog)
	filePathOutputLog := filepath.Join(basePath, "script", outputLog)

	// create log files
	_, err = os.Create(filePathErrorLog)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = os.Create(filePathOutputLog)
	if err != nil {
		log.Fatalln(err)
	}

	pt := PlistTemplate{
		OutputFileName:      "com.shortbread.client",
		PathToClientBinary:  filepath.Join(pwd, "client"),
		ShortbreadServerURL: os.Args[1],
		PathToErrorLog:      filePathErrorLog,
		PathToOutputLog:     filePathOutputLog,
		TimeInterval:        "300",
	}

	temp, err := template.ParseFiles(filePathTemplate)
	if err != nil {
		log.Fatalln(err)
	}

	var b []byte
	buf := bytes.NewBuffer(b)

	err = temp.Execute(buf, pt)
	if err != nil {
		log.Fatalln(err)
	}

	err = ioutil.WriteFile(filePathPlist, buf.Bytes(), 0644)
	if err != nil {
		log.Fatalln(err)
	}
}
