package main

import (
  "net/http"
  "sync"
  "path/filepath"
  "text/template"
)

type HandlRoute struct {
    Name        string
    Method      string
    Pattern     string
    Handler     http.Handler
}

type FuncRoute struct {
    Name        string
    Method      string
    Pattern     string
    HandlerFunc http.HandlerFunc
}
type AdminRoute struct {
    Name        string
    Method      string
    Pattern     string
    HandlerFunc http.HandlerFunc
}
type HandlRoutes []HandlRoute
type FuncRoutes  []FuncRoute
type AdminRoutes []AdminRoute

type templateHandler struct {
    once     sync.Once
    filename string
    templ    *template.Template
}

// ServeHTTP handles the HTTP request.
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r*http.Request) {
    t.once.Do (func() {
        t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
    })
    data := map[string]interface{}{
        "Host": r.Host,
    }
    t.templ.Execute(w, data)
}

// Define the routes served by our application.
var handlRoutes = HandlRoutes{
    HandlRoute{
        "Certs",
        "GET",
        "/certs",
        MustAuth(&templateHandler{filename: "create_cert.html"}),
    },
    HandlRoute{
        "Login",
        "GET",
        "/login",
        &templateHandler{filename: "login.html"},
    },
}
var funcRoutes = FuncRoutes{
    FuncRoute{
        "HandleRoot",
        "GET",
        "/",
        RootHandler,
    },
    FuncRoute{
        "HandleLogin",
        "GET",
        "/auth/login",
        LoginHandler,
    },
    FuncRoute{
        "SAMLCallback",
        "POST",
        "/v1/_saml_callback",
        AssertionHandler,
    },
    FuncRoute{
        "CreateCertificates",
        "POST",
        "/v1/createcert",
        TokenAuth(CreateCertHandler).(http.HandlerFunc),
    },
//    FuncRoute{
//        "AuthToken",
//        "POST",
//        "/v1/auth/login",
//        GetTokenHandler,
//    },
    FuncRoute{
        "ValidateToken",
        "GET",
        "/v1/auth/tokencheck",
        TokenAuth(OKHandler).(http.HandlerFunc),
    },
}
var adminRoutes = AdminRoutes{
    AdminRoute{
        "Ping",
        "GET",
        "/ping",
        PingHandler,
    },
    AdminRoute{
        "Healthcheck",
        "GET",
        "/healthcheck",
        HealthcheckHandler,
    },
}
