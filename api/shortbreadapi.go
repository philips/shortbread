// Package api provides access to the API to communicate with a centralized CA.
//
// See https://github.com/philips/shortbread
//
// Usage example:
//
//   import "code.google.com/p/google-api-go-client/api/v1"
//   ...
//   apiService, err := api.New(oauthHttpClient)
package api

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

const apiId = "api:v1"
const apiName = "api"
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

type CertificateAndPrivateKey struct {
	Cert string `json:"cert,omitempty"`

	PrivateKey string `json:"privateKey,omitempty"`
}

type CertificateInfo struct {
	// CertType: only accepts HOST or USER
	CertType string `json:"CertType,omitempty"`

	Key string `json:"Key,omitempty"`

	Permission *Permissions `json:"Permission,omitempty"`

	// PrivateKey: path of the private key on the CA server
	PrivateKey string `json:"PrivateKey,omitempty"`

	User string `json:"User,omitempty"`

	ValidAfter uint64 `json:"ValidAfter,omitempty,string"`

	ValidBefore uint64 `json:"ValidBefore,omitempty,string"`
}

type CertificatesWithKey struct {
	List []*CertificateAndPrivateKey `json:"list,omitempty"`
}

type Permissions struct {
	CriticalOptions []string `json:"criticalOptions,omitempty"`

	Extensions []string `json:"extensions,omitempty"`
}

type RevokeCertificate struct {
	Key string `json:"Key,omitempty"`

	User string `json:"User,omitempty"`
}

// method id "api.cert.getCerts":

type CertGetCertsCall struct {
	s         *Service
	publicKey string
	opt_      map[string]interface{}
}

// GetCerts:
func (r *CertService) GetCerts(publicKey string) *CertGetCertsCall {
	c := &CertGetCertsCall{s: r.s, opt_: make(map[string]interface{})}
	c.publicKey = publicKey
	return c
}

// PublicKey sets the optional parameter "publicKey":
func (c *CertGetCertsCall) PublicKey(publicKey string) *CertGetCertsCall {
	c.opt_["publicKey"] = publicKey
	return c
}

func (c *CertGetCertsCall) Do() (*CertificatesWithKey, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", "json")
	if v, ok := c.opt_["publicKey"]; ok {
		params.Set("publicKey", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "getcerts/{publicKey}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"publicKey": c.publicKey,
	})
	req.Header.Set("User-Agent", "google-api-go-client/0.5")
	res, err := c.s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	var ret *CertificatesWithKey
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "httpMethod": "GET",
	//   "id": "api.cert.getCerts",
	//   "parameterOrder": [
	//     "publicKey"
	//   ],
	//   "parameters": {
	//     "publicKey": {
	//       "default": "",
	//       "format": "bytes",
	//       "location": "path",
	//       "required": "true",
	//       "type": "string"
	//     }
	//   },
	//   "path": "getcerts/{publicKey}",
	//   "response": {
	//     "$ref": "CertificatesWithKey"
	//   }
	// }

}

// method id "api.cert.revoke":

type CertRevokeCall struct {
	s                 *Service
	revokecertificate *RevokeCertificate
	opt_              map[string]interface{}
}

// Revoke:
func (r *CertService) Revoke(revokecertificate *RevokeCertificate) *CertRevokeCall {
	c := &CertRevokeCall{s: r.s, opt_: make(map[string]interface{})}
	c.revokecertificate = revokecertificate
	return c
}

func (c *CertRevokeCall) Do() error {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.revokecertificate)
	if err != nil {
		return err
	}
	ctype := "application/json"
	params := make(url.Values)
	params.Set("alt", "json")
	urls := googleapi.ResolveRelative(c.s.BasePath, "revoke")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("PUT", urls, body)
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
	//   "httpMethod": "PUT",
	//   "id": "api.cert.revoke",
	//   "path": "revoke",
	//   "request": {
	//     "$ref": "RevokeCertificate",
	//     "parameterName": "revokeCertParams"
	//   }
	// }

}

// method id "api.cert.sign":

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
	//   "id": "api.cert.sign",
	//   "path": "sign",
	//   "request": {
	//     "$ref": "CertificateInfo",
	//     "parameterName": "certParams"
	//   }
	// }

}
