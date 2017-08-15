package main

import (
    "net/http"

    "github.com/gorilla/mux"
)

// NewRouter creates a custom instance of mux.Router which describes a route served by the server
func NewRouter() *mux.Router {

    router := mux.NewRouter().StrictSlash(true)
    // Set up static files here as they must be done on the route itself, not
    // passed to it like functions/handlers
    router.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("./js"))))
    // set up the routes
    for _, route := range handlRoutes {
        var handler http.Handler

        handler = route.Handler
        handler = HTTPLogger(handler, route.Name)

        router.
            Methods(route.Method).
            Path(route.Pattern).
            Name(route.Name).
            Handler(handler)

    }
    for _, route := range funcRoutes {
        var handler http.Handler

        handler = route.HandlerFunc
        handler = HTTPLogger(handler, route.Name)

        router.
            Methods(route.Method).
            Path(route.Pattern).
            Name(route.Name).
            Handler(handler)

    }

    return router
}

// NewAdminRouter creates a custom instance of mux.Router which describes a
// route served by the server
// These routes are intended to be served on an admin port (e.g. 8081 instead
// of 8080) and should be created with that in mind
func NewAdminRouter() *mux.Router {

    router := mux.NewRouter().StrictSlash(true)
    for _, route := range adminRoutes {
        var handler http.Handler

        handler = route.HandlerFunc
        handler = HTTPLogger(handler, route.Name)

        router.
            Methods(route.Method).
            Path(route.Pattern).
            Name(route.Name).
            Handler(handler)

    }

    return router
}
