package main

import (
    "gopkg.in/yaml.v2"
    "crypto/sha256"
    "encoding/base64"
    "time"
    "strings"
    "net/http"
    "encoding/json"
    "bytes"
    "io/ioutil"
    "os"
    "fmt"
    "errors"
    "sync"
)

const (
    LemurUserEnv      = "LEMUR_USER"
    LemurPasswordEnv  = "LEMUR_PASS"
    LemurUrl          = "https://lemur.example.com"
    LemurApiVersion   = "/api/1"
    AuthorityUri      = "/authorities"
    DestinationsUri   = "/destinations"
    CertificatesUri   = "/certificates"
    AuthorizeUri      = "/auth/login"
)

type authToken struct {
    Token string `json:"token"`
}

type LemurRequester struct {
    Token     string
    Client    *http.Client
}

type Authority struct {
    Name  string
    Id    int
}

type certJsonRequest struct {
    Authority           string                                `yaml:"authority"           json:"authority"`
    CommonName          string                                `yaml:"commonName"          json:"commonName"`
    Email               string                                `yaml:"owner"               json:"owner"`
    StartDate           string                                `yaml:"validityStart"       json:"validityStart"`
    EndDate             string                                `yaml:"validityEnd"         json:"validityEnd"`
}

type certManifest struct {
    Authority           map[string]string                     `yaml:"authority"           json:"authority"`
    CommonName          string                                `yaml:"commonName"          json:"commonName"`
    Email               string                                `yaml:"owner"               json:"owner"`
    StartDate           string                                `yaml:"validityStart"       json:"validityStart"`
    EndDate             string                                `yaml:"validityEnd"         json:"validityEnd"`
    Description         string                                `yaml:"description"         json:"description"`
    Country             string                                `yaml:"country"             json:"country"`
    State               string                                `yaml:"state"               json:"state"`
    Location            string                                `yaml:"location"            json:"location"`
    Organization        string                                `yaml:"organization"        json:"organization"`
    OrganizationalUnit  string                                `yaml:"organizationalUnit"  json:"organizationalUnit"`
    Active              bool                                  `yaml:"active"              json:"active"`
    Extensions          map[string]map[string]map[string]bool `yaml:"extensions"          json:"extensions"`
    Once                sync.Once                             `yaml:"-"                   json:"-"`
}

type certChainPubKey struct {
    Chain             string `yaml:"chain"      json:"chain"`
    PublicCertificate string `yaml:"pubcert"    json:"pubcert"`
    PrivateKey        string `yaml:"privatekey" json:"privatekey"`
}

func NewCertManifest(authority, commonName, email, start, end, rbacgroup string) (*certManifest) {
    var extensions = map[string]map[string]map[string]bool{
        "extensions": {"keyUsage": {"isCritical": true,
                                    "useDigitalSignature": true},
                       "extendedKeyUsage": {"isCritical": true,
                                            "clientAuth": true},
                       "subjectKeyIdentifier": {"isCritical": false,
                                                "includeSKI": true}}}
    authObj := map[string]string{"name": authority}
    man := certManifest{Authority: authObj,
                        CommonName: commonName,
                        Email: email,
                        StartDate: start,
                        EndDate: end,
                        Description: "Temporary Client Certificate (2 weeks)",
                        Country: "US",
                        State: "OR",
                        Location: "Portland",
                        Organization: rbacgroup,
                        OrganizationalUnit: rbacgroup,
                        Active: true,
                        Extensions: extensions}
    man.makeDigest() // make sure this is the only call to makeDigest
    return &man
}

// makeDigest hashes the description of the certificate manifest, guaranteeing
// that it is unique and therefore searchable. Useful since tracking IDs is
// difficult.
// Called on a certManifest pointer
// Returns nothing
func (c *certManifest) makeDigest(){
    c.Once.Do(func() {
        bytesBuf, _ := yaml.Marshal(*c)
        hasher := sha256.New()
        hasher.Write(bytesBuf)
        desc := fmt.Sprintf("%s:%s",
                            base64.URLEncoding.EncodeToString(hasher.Sum(nil)),
                            c.Description)
        c.Description = desc
    })
}

func (l *LemurRequester) doCheckRequest(r *http.Request) ([]byte, error) {
    r.Header.Set("Content-Type", "application/json")
    r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", l.Token))
    response, err := l.Client.Do(r)
    if err != nil {
        Logs.Errorf("Unable to make HTTP Client Request to '%+v'\nError: %+v\n", r, err)
        return nil, err
    }
    nonErrorHttpHundreds := map[int]bool{2: true, 3: true}
    if !nonErrorHttpHundreds[response.StatusCode / 100] {
        defer response.Body.Close()
        body, _ := ioutil.ReadAll(response.Body)
        return nil, fmt.Errorf("Bad response (%+v) from request (%+v), response body: '%+v'", response.StatusCode, r.URL, string(body))
    }
    bufferBody, err := ioutil.ReadAll(response.Body)
    if err != nil {
        Logs.Errorf("Unable to read response body from request'%+v'\nError: %+v\n", r, err)
        return nil, err
    }
    return bufferBody, nil
}


func (l *LemurRequester) getAuthToken() (error) {
    if len(l.Token) > 0 {
        return nil
    }
    user := os.Getenv(LemurUserEnv)
    if len(user) == 0 {
        return errors.New("LEMUR_USER|LEMUR_PASS environment variable(s) not set. Cannot continue.")
    }
    pass := os.Getenv(LemurPasswordEnv)
    l.Client = &http.Client{Timeout: time.Second * 30}
    data := map[string]string{"username": user, "password": pass}
    jsonData, _ := json.Marshal(&data)
    address := []string{LemurUrl, LemurApiVersion, AuthorizeUri}
    authReq, err := http.NewRequest("POST",
                                    strings.Join(address, ""),
                                    bytes.NewBuffer(jsonData))
    bufferBody, err := l.doCheckRequest(authReq)
    if err != nil {
        return err
    }
    var tokenBody authToken
    decoder := json.NewDecoder(bytes.NewReader(bufferBody))
    decoder.UseNumber()
    if err := decoder.Decode(&tokenBody); err != nil {
        return err
    }
    l.Token = tokenBody.Token
    return nil
}

// ValidateCert checks the manifest for existing certificates and creates one
// if none are found
// Called on a LemurRequester pointer
// Takes a certManifest pointer as argument
// Returns a certChainPubKey pointer and an error
func (l *LemurRequester) ValidateCert(c *certManifest) (*certChainPubKey, error) {
    if err := l.getAuthToken(); err != nil {
        Logs.Errorf("Error ensuring auth token in ValidateCert\nError: %+v\n", err)
        return nil, err
    }
    address := []string{LemurUrl, LemurApiVersion, CertificatesUri}
    getCertReq, err := http.NewRequest("GET",
                                       strings.Join(address, ""),
                                       nil)
    query := getCertReq.URL.Query()
    query.Add("sortBy", "date_created")
    query.Add("sortDir", "desc")
    query.Add("filter", fmt.Sprintf("%s;%s", "description", c.Description))
    getCertReq.URL.RawQuery = query.Encode()
    Logs.Infof("Making request to url %+v", getCertReq.URL)
    bufferBody, err := l.doCheckRequest(getCertReq)
    if err != nil {
        return nil, err
    }
    var certificates interface{}
    decoder := json.NewDecoder(bytes.NewReader(bufferBody))
    decoder.UseNumber()
    err = decoder.Decode(&certificates)
    if err != nil {
        Logs.Errorf("Unable to Unmarshal certificates data to json: %+v", err)
        return nil, err
    }
    certsMap := certificates.(map[string]interface{})
    numCerts, err := certsMap["total"].(json.Number).Int64()
    if err != nil {
        return nil, err
    }
    var newestCert map[string]interface{}
    if numCerts < 1 {
        Logs.Infof("No cert matching manifest exists yet. Trying to create new cert.\n")
        newestCert, err = l.createCert(c)
        if err != nil {
            Logs.Errorf("Unable to create new certificate: %+v\n", err)
            return nil, err
        }
    } else {
        // else certs >= 1, so just use the newest one
        Logs.Infof("At least one cert matching manifest exist. Returning newest instance.\n")
        items := certsMap["items"].([]interface{})
        newestCert = items[0].(map[string]interface{})
    }
    chain := newestCert["chain"].(string)
    body := newestCert["body"].(string)
    certId, _ := newestCert["id"].(json.Number).Int64()
    key, err := l.getCertKeyById(certId)
    if err != nil {
        Logs.Errorf("Unable to get certificate key by id: %+v\n", err)
        return nil, err
    }
    return &certChainPubKey{Chain: chain,
                            PublicCertificate: body,
                            PrivateKey: key}, nil
}

// createCert creates a new certificate
// Called on a LemurRequester pointer
// Takes a certManifest pointer as argument
// Returns a map of string to interface and an error
func (l *LemurRequester) createCert(c *certManifest) (map[string]interface{}, error) {
    if err := l.getAuthToken(); err != nil {
        Logs.Errorf("Error ensuring auth token in getCertKeyById()\nError: %+v\n", err)
        return nil, err
    }
    jsonData, _ := json.Marshal(*c)
    address := []string{LemurUrl, LemurApiVersion, CertificatesUri}
    Logs.Infof("Trying to create new certificate with address %+v and data %+v", address, string(jsonData))
    certReq, err := http.NewRequest("POST",
                                    strings.Join(address, ""),
                                    bytes.NewReader(jsonData))
    bufferBody, err := l.doCheckRequest(certReq)
    if err != nil {
        Logs.Errorf("Unable to POST request for new certificate: %+v", err)
        return nil, err
    }
    var certificates interface{}
    decoder := json.NewDecoder(bytes.NewReader(bufferBody))
    decoder.UseNumber()
    if err := decoder.Decode(&certificates); err != nil {
        Logs.Errorf("Unable to Unmarshal new certificate data to json\nError: %+v\n", err)
        return nil, err
    }
    certsMap := certificates.(map[string]interface{})
    return certsMap, nil
}

// getCertKeyById gets the private key for a certificate given that cert's id
// Called on a LemurRequester pointer
// Takes an int as argument
// Returns a string and an error
func (l *LemurRequester) getCertKeyById(id int64) (string, error) {
    if err := l.getAuthToken(); err != nil {
        Logs.Errorf("Error ensuring auth token in getCertKeyById()\nError: %+v\n", err)
        return "", err
    }
    address := []string{LemurUrl,
                        LemurApiVersion,
                        CertificatesUri,
                        fmt.Sprintf("/%d",id),
                        "/key"}
    getCertKeyReq, err := http.NewRequest("GET",
                                          strings.Join(address, ""),
                                          nil)
    bufferBody, err := l.doCheckRequest(getCertKeyReq)
    if err != nil {
        return "", err
    }
    var key interface{}
    decoder := json.NewDecoder(bytes.NewReader(bufferBody))
    decoder.UseNumber()
    if err := decoder.Decode(&key); err != nil {
        Logs.Errorf("Unable to Unmarshal certificate key to json. Error: %+v\n", err)
        return "", err
    }
    keyMap := key.(map[string]interface{})
    certKey := keyMap["key"].(string)
    return certKey, nil
}
