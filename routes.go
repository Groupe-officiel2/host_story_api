// routes.go

package main

import "net/http"

// RegisterRoutes sets up all the HTTP routes for the application
func RegisterRoutes() {
	http.HandleFunc("/public", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Public endpoint - no authentication required"))
	})

	http.HandleFunc("/protected", WithJWTAuth(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Protected endpoint - valid JWT required"))
	}))

	http.HandleFunc("/template", WithJWTAuth(CreateTemplateContainer))
	http.HandleFunc("/toggle", WithJWTAuth(ToggleHandler))
    http.HandleFunc("/players", WithJWTAuth(GetPlayers))
}
