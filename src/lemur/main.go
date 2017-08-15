package main

import (
    "crypto/tls"
    "net/http"
    "net"
    "time"
    "github.com/DataDog/datadog-go/statsd"
)
//"net/url"
var LemurCertsStatsd *statsd.Client
var LemurHttpStatsd *statsd.Client
var OktaProvider *oktaProvider
var secretKey *authSecret
var Flags *flagOptArgs

type tcpKeepAliveListener struct {
	*net.TCPListener
}
// Satisfy the TCPListener interface
func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

// templ represents a single template
func main() {
    // Set up our progrumme
    Flags = GetFlags()
    Logs.Infof("Got flags: %+v\n", Flags)

    // Set up our app/auth secret
    Logs.Infof("Creating apptoken secret...")
    secretKey = NewTokenSecret()
    Logs.Infof("Created apptoken secret!")

    // Set up a statsd client for certs metrics
    LemurCertsStatsd = NewCertsStatsd(*Flags.StatsdHost, *Flags.StatsdPort)
    Logs.Infof("Set up statsd client pointed at %s:%s with prefix %s",
               *Flags.StatsdHost,
               *Flags.StatsdPort,
               "lemur.certs")
    defer LemurCertsStatsd.Close() // Unfortunately this must be done in main()
    Logs.Infof("lemur.certs statsd client established, close deferred")

    // Set up a statsd client for http metrics
    LemurHttpStatsd = NewHttpStatsd(*Flags.StatsdHost, *Flags.StatsdPort)
    Logs.Infof("Set up statsd client pointed at %s:%s with prefix %s",
               *Flags.StatsdHost,
               *Flags.StatsdPort,
               "lemur.http")
    defer LemurHttpStatsd.Close() // Unfortunately this must be done in main()
    Logs.Infof("lemur.http statsd client established")

    // Set up the okta provider so we can auth
    OktaProvider = NewOktaProvider(&*Flags.Config)

    // Set up the routes for the web server
    router := NewRouter()

    // Set up the certificates/certificate reloader for the web server(s)
    kpr, err := NewKeypairReloader("server.crt", "server.key", &*Flags.Config)
    if err != nil {
        Logs.Errorf("Unable to create NewKeypairReloader.")
        panic(err)
    }

    go func() {
        // Set up the web server
        Logs.Infof("Starting web server.")
        srv := &http.Server{
            Addr: *Flags.HostPort,
            Handler: router}
        srv.TLSConfig = &tls.Config{
            GetCertificate: kpr.GetCertificateFunc(),
        }
        ln, err := net.Listen("tcp", *Flags.HostPort)
        if err != nil {
            Logs.Errorf("Could not start tcp listener: %+v", err)
            panic(err)
        }
        tlsListener := tls.NewListener(tcpKeepAliveListener{ln.(*net.TCPListener)}, srv.TLSConfig)
        srv.Serve(tlsListener)
//        if err := http.ListenAndServe(*Flags.HostPort, router); err != nil {
//          Logs.Errorf("%+v", err)
//        }
    }()

    // Set up admin routes
    adminRouter := NewAdminRouter()
    //Set up the admin web server
    Logs.Infof("Starting admin interface.")
    adminSrv := &http.Server{
        Addr: *Flags.AdminPort,
        Handler: adminRouter}
    adminSrv.TLSConfig = &tls.Config{
        GetCertificate: kpr.GetCertificateFunc(),
    }
    ln, err := net.Listen("tcp", *Flags.AdminPort)
    if err != nil {
        Logs.Errorf("Could not start tcp listener for admin interface: %+v", err)
        panic(err)
    }
    tlsListener := tls.NewListener(tcpKeepAliveListener{ln.(*net.TCPListener)}, adminSrv.TLSConfig)
    adminSrv.Serve(tlsListener)
//    if err := http.ListenAndServe(*Flags.AdminPort, adminRouter); err != nil {
//        Logs.Errorf("%+v", err)
//    }
}

