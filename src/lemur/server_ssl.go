package main

import (
    "crypto/tls"
    "sync"
    "time"
    "io/ioutil"
)

type keypairReloader struct {
    config   *InstanceConfig
    certMu   sync.RWMutex
    cert     *tls.Certificate
    certPath string
    keyPath  string
}

// writeFile writes data to path.
func writeFile(path, data string) {
    err := ioutil.WriteFile(path, []byte(data), 0600)
    if err != nil {
        Logs.Errorf("Unable to write certifcate to file: %+v", err)
        panic(err)
    }
}

// NewKeypairReloader attempts to set up server certificates on startup and
// then checks once per day to see if it's still valid. If not, try to reload.
func NewKeypairReloader(certPath, keyPath string, config *InstanceConfig) (*keypairReloader, error) {
    result := &keypairReloader{
        certPath: certPath,
        keyPath:  keyPath,
        config:   config,
    }
    // This should happen once at server startup
    if err := result.maybeReload(); err != nil {
        Logs.Errorf("Unable to initialize server certificate! This should not happen!")
        return nil, err
    }
    go func() {
        for {
            // Check once per day to see if our certs are OK
            time.Sleep(24 * time.Hour)
            cert, err := tls.LoadX509KeyPair(certPath, keyPath)
            if err != nil {
                Logs.Errorf("Unable to load certificates. It is likely that certificates will never be loaded!")
                panic(err)
            }
            // Give ourselves a 24 hour window to refresh
            refresh_soon := time.Now().Add(time.Duration(24) * time.Hour)
            // Cert is invalid or will be soon
            if time.Now().Before(cert.Leaf.NotBefore) || refresh_soon.After(cert.Leaf.NotAfter) {
                Logs.Infof("Server certificate is or will soon be out-of-date. Refreshing.")
                if err := result.maybeReload(); err != nil {
                    Logs.Errorf("Keeping old TLS certificate because the new one could not be loaded: %v", err)
                }
            }
        }
    }()
    return result, nil
}

// maybeReload requests a new certificate from our builtin broker handler and
// writes it to disk so the tls library can load it into the web server
func (kpr *keypairReloader) maybeReload() error {
    // Start every cert at the beginning of the month
    // As of this writing, SSL certs are expected to be generated via Let's
    // Encrypt authority, which has a rate limit of roughly 5 certs a week. Too
    // many restarts of the service in a week will rate-limit the application,
    // so reduce requests to once per month.
    start := time.Date(time.Now().Year(), time.Now().Month(), 1, 1, 1, 1, 1, time.UTC)
    end := start.Add(time.Duration(24) * time.Hour * 730) // two years
    man := NewCertManifest(kpr.config.CertAuthority,
                           kpr.config.CommonName,
                           kpr.config.EmailAddress,
                           start.Format("2006-01-02"),
                           end.Format("2006-01-02"),
                           kpr.config.CertOrg)
    lemurReq := &LemurRequester{}
    Logs.Infof("Requesting server certificate from broker.")
    chainCertKey, err := lemurReq.ValidateCert(man)
    if err != nil {
        LemurCertsStatsd.Incr("errors", nil, 1)
        Logs.Errorf("Requesting server certificate failed: %+v", err)
        return err
    }
    writeFile(kpr.certPath, chainCertKey.PublicCertificate)
    writeFile(kpr.keyPath, chainCertKey.PrivateKey)
    newCert, err := tls.LoadX509KeyPair(kpr.certPath, kpr.keyPath)
    if err != nil {
        return err
    }
    kpr.certMu.Lock()
    defer kpr.certMu.Unlock()
    kpr.cert = &newCert
    return nil
}

// GetCertificateFunc allows a webserver's tls config to update certificates
func (kpr *keypairReloader) GetCertificateFunc() func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
    return func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
        kpr.certMu.RLock()
        defer kpr.certMu.RUnlock()
        return kpr.cert, nil
    }
}
