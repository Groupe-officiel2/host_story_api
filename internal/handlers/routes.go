package handlers

import (
	"host_story_api/internal/auth"
	"net/http"
)

// RegisterRoutes sets up all the HTTP routes for the application
func RegisterRoutes() {
	http.HandleFunc("/public", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Public endpoint - no authentication required"))
	})

	http.HandleFunc("/protected", auth.WithJWTAuth(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Protected endpoint - valid JWT required"))
	}))

	http.HandleFunc("/template", auth.WithJWTAuth(CreateTemplateContainer))
	http.HandleFunc("/toggle", auth.WithJWTAuth(ToggleHandler))
	http.HandleFunc("/players", auth.WithJWTAuth(GetPlayers))
	http.HandleFunc("/servers", auth.WithJWTAuth(GetServers))
}
