package main

import "net/http"

// RegisterRoutes sets up all the HTTP routes for the application
func RegisterRoutes(apiKey string) {
	http.HandleFunc("/template", WithAPIKey(apiKey, CreateTemplateContainer))
	http.HandleFunc("/toggle", WithAPIKey(apiKey, ToggleHandler))
}
