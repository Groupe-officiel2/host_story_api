package main

import (
	"fmt"
	"net/http"
	"sync"
	"github.com/joho/godotenv"
)

const port = ":8082"

var (
	serverCounter int
	serverMutex   sync.Mutex
)

func main() {
	godotenv.Load()

	RegisterRoutes()

	fmt.Printf("(http://localhost%s)\n", port)
	http.ListenAndServe(port, nil)
}
