package main

import (
    "net/http"
    "time"
)

// HTTPLogger decorates http.Handler instances such that it logs requests made to the server
func HTTPLogger(inner http.Handler, name string) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        inner.ServeHTTP(w, r)

        Logs.Infof(
            "%s\t%s\t%s\t%s",
            r.Method,
            r.RequestURI,
            name,
            time.Since(start),
        )
    })
}
