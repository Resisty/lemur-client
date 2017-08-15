package main

import (
  "net/http"
  "fmt"
  "encoding/json"
)


type authHandler struct {
    next http.Handler
}

type apiAuthHandler struct {
    next http.HandlerFunc
}

type authJsonRequest struct {
    UserName string `json:"username"`
    Password string `json:"password"`
}

func (h *authHandler) ServeHTTP (w http.ResponseWriter, r *http.Request) {
    var cookie, err = r.Cookie("auth")
    if  err == http.ErrNoCookie || cookie.Value == "" {
        w.Header().Set("Location", "/login")
        w.WriteHeader(http.StatusTemporaryRedirect)
        return
    }
    token := cookie.Value
    if _, err := secretKey.ValidateToken(token) ; err != nil {
        w.Header().Set("Location", "/login")
        w.WriteHeader(http.StatusTemporaryRedirect)
        return
    }
    h.next.ServeHTTP(w, r)
}

func (h *apiAuthHandler) ServeHTTP (w http.ResponseWriter, r *http.Request) {
    var token = r.Header.Get("Authorization")
    if _, err := secretKey.ValidateToken(token) ; err != nil {
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte("401 - Bad token."))
        return
    } else {
        // success - call the next handler
        h.next.ServeHTTP(w, r)
    }
}

func MustAuth(handler http.Handler) http.Handler {
    return &authHandler{next: handler}
}

func TokenAuth(h http.HandlerFunc) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var token = r.Header.Get("Authorization")
        if _, err := secretKey.ValidateToken(token) ; err != nil {
            w.WriteHeader(http.StatusUnauthorized)
            w.Write([]byte("401 - Bad token."))
            return
        }
        h.ServeHTTP(w, r)
    })
}

func OKHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("200 - Well met!"))
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Location", "/certs")
    w.WriteHeader(http.StatusTemporaryRedirect)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Location", OktaProvider.ServiceProvider.IdentityProviderSSOURL)
    w.WriteHeader(http.StatusTemporaryRedirect)
}

func AssertionHandler (w http.ResponseWriter, r *http.Request) {
    err := r.ParseForm()
    if err != nil {
        LemurHttpStatsd.Incr("error", nil, 1)
        Logs.Errorf("Unable to understand form: %+v", err)
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte("400 - Unable to understand form."))
        return
    }
    assertionInfo, err := OktaProvider.ServiceProvider.RetrieveAssertionInfo(r.FormValue("SAMLResponse"))
    if err != nil {
        LemurHttpStatsd.Incr("error", nil, 1)
        Logs.Errorf("Unable to understand assertion information: %+v", err)
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte("400 - Unable to understand assertion information."))
        return
    }
    LemurCertsStatsd.Incr("authenticate", nil, 1)
    data, err := secretKey.MakeToken(assertionInfo.Values["username"].Values[0].Value,
                                     assertionInfo.Values["rbac"].Values[0].Value)
    if err != nil {
        LemurHttpStatsd.Incr("error", nil, 1)
        Logs.Errorf("Unable to generate authentication token: %+v", err)
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("500 - Unable to generate authentication token."))
        return
    }
    LemurHttpStatsd.Incr("tokens", nil, 1)
    http.SetCookie(w, &http.Cookie{
        Name: "auth",
        Value: data.(map[string]string)["token"],
        Path: "/"})
    w.Header().Set("Location", "/certs")
    // Use http.StatusSeeOther to get around annoying RPG problem
    // https://en.wikipedia.org/wiki/Post/Redirect/Get
    http.Redirect(w, r, "/certs", http.StatusSeeOther)
}

func CreateCertHandler (w http.ResponseWriter, r *http.Request) {
    var token = r.Header.Get("Authorization")
    claims := secretKey.GetClaims(token)
    if _, ok := claims["rbac"]; !ok {
		LemurCertsStatsd.Incr("errors", nil, 1)
		Logs.Errorf("%+v", ok)
        w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, ok)
        return
    }
    rbacGroup := claims["rbac"].(string)
    decoder := json.NewDecoder(r.Body)
    var certReq certJsonRequest
    err := decoder.Decode(&certReq)
    if err != nil {
		LemurCertsStatsd.Incr("errors", nil, 1)
		Logs.Errorf("%+v", err)
	}
	LemurCertsStatsd.Incr("requests", nil, 1)
    defer r.Body.Close()
	manifest := NewCertManifest(certReq.Authority,
								certReq.CommonName,
								certReq.Email,
								certReq.StartDate,
								certReq.EndDate,
                                rbacGroup)
	lemurReq := &LemurRequester{}
	chainCertKey, err := lemurReq.ValidateCert(manifest)
	if err != nil {
		LemurCertsStatsd.Incr("errors", nil, 1)
		Logs.Errorf("%+v", err)
        w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
	} else {
		LemurCertsStatsd.Incr("issued", nil, 1)
		output, _ := json.Marshal(chainCertKey)
        w.Header().Set("Content-Type", "application/json")
		w.Write(output)
	}
}

func GetTokenHandler (w http.ResponseWriter, r *http.Request) {
    decoder := json.NewDecoder(r.Body)
    var authReq authJsonRequest
    err := decoder.Decode(&authReq)
    if err != nil {
        LemurHttpStatsd.Incr("errors", nil, 1)
        Logs.Errorf("%+v", err)
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte("400 - Unable to understand request."))
        return
    }
    response, err := OktaProvider.OktaApiAuth(authReq)
    if err != nil {
        LemurHttpStatsd.Incr("errors", nil, 1)
        Logs.Errorf("%+v", err)
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("500 - Unable to send request to Idp API auth url."))
        return
    } else if response.StatusCode >= 400 {
        w.WriteHeader(response.StatusCode)
        w.Write([]byte(fmt.Sprintf("%d - authentication failure", response.StatusCode)))
        return
    }
    Logs.Infof("Made a login attempt against %s, status code %d", OktaProvider.Config.IdpApiAuthUrl, response.StatusCode)
    data, err := secretKey.MakeToken(authReq.UserName, "")
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("500 - Unable to generate authentication token."))
        return
    }
    LemurHttpStatsd.Incr("tokens", nil, 1)
    output, _ := json.Marshal(data)
    w.WriteHeader(http.StatusOK)
    w.Header().Set("Content-Type", "application/json")
    w.Write(output)
}

// PingHandler always returns 200 OK
func PingHandler (w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("SYNACK"))
    return
}

// HealthcheckHandler always returns 200 OK
// At the time of this writing, there is no scenario in which normal program
// execution will continue but also be unusable; any error preventing use will
// also panic()
func HealthcheckHandler (w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`{"application": {"healthy": true} }`))
    return
}
