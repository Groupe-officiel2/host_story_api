package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

const port = ":8080"
const stateFile = "server_state.json"

var (
	serverCounter     int
	serverMutex       sync.Mutex
	baseHostPort      = 42720
	baseContainerPort = 42420
)

func main() {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY is required")
	}

	RegisterRoutes(apiKey)

	fmt.Printf("(http://localhost:8080) - Server is running on port %s\n", port)
	http.ListenAndServe(port, nil)
}
