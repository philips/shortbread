// Package client provides access to the API to communicate with a centralized CA.
//
// See https://github.com/philips/shortbread
//
// Usage example:
//
//   import "code.google.com/p/google-api-go-client/client/v1"
//   ...
//   clientService, err := client.New(oauthHttpClient)
package client

import (
	"bytes"
	"code.google.com/p/google-api-go-client/googleapi"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Always reference these packages, just in case the auto-generated code
// below doesn't.
var _ = bytes.NewBuffer
var _ = strconv.Itoa
var _ = fmt.Sprintf
var _ = json.NewDecoder
var _ = io.Copy
var _ = url.Parse
var _ = googleapi.Version
var _ = errors.New
var _ = strings.Replace

const apiId = "client:v1"
const apiName = "client"
const apiVersion = "v1"
const basePath = "https://www.example.com/v1/"

func New(client *http.Client) (*Service, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}
	s := &Service{client: client, BasePath: basePath}
	s.Cert = NewCertService(s)
	return s, nil
}

type Service struct {
	client   *http.Client
	BasePath string // API endpoint base URL

	Cert *CertService
}

func NewCertService(s *Service) *CertService {
	rs := &CertService{s: s}
	return rs
}

type CertService struct {
	s *Service
}

type CertificateInfo struct {
	// CertType: only accepts HOST or USER
	CertType string `json:"CertType,omitempty"`

	Key string `json:"Key,omitempty"`

	Permission *Permissions `json:"Permission,omitempty"`

	// PrivateKey: path of the private key on the CA server
	PrivateKey string `json:"PrivateKey,omitempty"`

	User string `json:"User,omitempty"`
}

type Permissions struct {
	CriticalOptions []string `json:"criticalOptions,omitempty"`

	Extensions []string `json:"extensions,omitempty"`
}

// method id "client.cert.sign":

type CertSignCall struct {
	s               *Service
	certificateinfo *CertificateInfo
	opt_            map[string]interface{}
}

// Sign:
func (r *CertService) Sign(certificateinfo *CertificateInfo) *CertSignCall {
	c := &CertSignCall{s: r.s, opt_: make(map[string]interface{})}
	c.certificateinfo = certificateinfo
	return c
}

func (c *CertSignCall) Do() error {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.certificateinfo)
	if err != nil {
		return err
	}
	ctype := "application/json"
	params := make(url.Values)
	params.Set("alt", "json")
	urls := googleapi.ResolveRelative(c.s.BasePath, "sign")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.SetOpaque(req.URL)
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", "google-api-go-client/0.5")
	res, err := c.s.client.Do(req)
	if err != nil {
		return err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return err
	}
	return nil
	// {
	//   "httpMethod": "POST",
	//   "id": "client.cert.sign",
	//   "path": "sign",
	//   "request": {
	//     "$ref": "CertificateInfo",
	//     "parameterName": "certParams"
	//   }
	// }

}
