package main

import (
  "github.com/russellhaering/gosaml2"
  "github.com/russellhaering/goxmldsig"
  "github.com/russellhaering/gosaml2/types"
  "crypto/x509"
  "io/ioutil"
  "net/http"
  "time"
  "encoding/json"
  "encoding/xml"
  "encoding/base64"
  "bytes"
)

var configs instanceConfig
type instanceConfig struct {
    IdpMetadata    string `yaml:"idp_metadata"`
    IdpIssuer      string `yaml:"idp_issuer"`
    IdpApiAuthUrl  string `yaml:"idp_api_auth_url"`
    SamlCallback   string `yaml:"saml_callback"`
    AudienceURI    string `yaml:"audience_uri"`
}

type oktaProvider struct {
    ServiceProvider *saml2.SAMLServiceProvider
    Config          *InstanceConfig
}

func NewOktaProvider(configs *InstanceConfig) *oktaProvider {
    res, err := http.Get(configs.IdpMetadata)
	if err != nil {
        Logs.Errorf("%+v",err)
		panic(err)
	}
	rawMetadata, err := ioutil.ReadAll(res.Body)
	if err != nil {
        Logs.Errorf("%+v",err)
		panic(err)
	}
	metadata := &types.EntityDescriptor{}
	err = xml.Unmarshal(rawMetadata, metadata)
	if err != nil {
        Logs.Errorf("%+v",err)
		panic(err)
	}
	certStore := dsig.MemoryX509CertificateStore{
		Roots: []*x509.Certificate{},
	}
	for _, kd := range metadata.IDPSSODescriptor.KeyDescriptors {
		certData, err := base64.StdEncoding.DecodeString(kd.KeyInfo.X509Data.X509Certificate.Data)
		if err != nil {
            Logs.Errorf("%+v",err)
			panic(err)
		}
		idpCert, err := x509.ParseCertificate(certData)
		if err != nil {
            Logs.Errorf("%+v",err)
			panic(err)
		}
		certStore.Roots = append(certStore.Roots, idpCert)
	}
    // We sign the AuthnRequest with a random key because Okta doesn't seem
    // to verify these.
    randomKeyStore := dsig.RandomKeyStoreForTest()
    sp := &saml2.SAMLServiceProvider{
        IdentityProviderSSOURL:      metadata.IDPSSODescriptor.SingleSignOnService.Location,
        IdentityProviderIssuer:      configs.IdpIssuer,
        AssertionConsumerServiceURL: configs.SamlCallback,
        SignAuthnRequests:           true,
        AudienceURI:                 "123",
        IDPCertificateStore:         &certStore,
        SPKeyStore:                  randomKeyStore,
    }
    return &oktaProvider{ServiceProvider: sp, Config: configs}
}

func (o *oktaProvider) OktaApiAuth (jsonData authJsonRequest) (*http.Response, error) {
    client := &http.Client{Timeout: time.Second * 30}
    jsonString, _ := json.Marshal(jsonData)
    authReq, err := http.NewRequest("POST",
                                    o.Config.IdpApiAuthUrl,
                                    bytes.NewBuffer(jsonString))
    authReq.Header.Set("Content-Type", "application/json")
    authReq.Header.Add("Accept", "application/json")
    client = &http.Client{Timeout: time.Second * 30}
    response, err := client.Do(authReq)
    if err != nil {
        Logs.Errorf("Unable to make HTTP Client request to '%+v'\nError: %+v\n", o.Config.IdpApiAuthUrl, err)
        return nil, err
    }
    return response, nil
}
