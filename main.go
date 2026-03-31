package main

import (
	"fmt"
	"net/http"
	"sync"
)

const port = ":8082"

var (
	serverCounter     int
	serverMutex       sync.Mutex
	baseHostPort      = 42720
	baseContainerPort = 42420
)

func main() {
	RegisterRoutes()

	fmt.Printf("(http://localhost:8082) - Server is running on port %s\n", port)
	http.ListenAndServe(port, nil)
}
