package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"syscall"

	"github.com/joho/godotenv"

	"host_story_api/internal/handlers"
)

const port = ":8082"

func main() {
	err := godotenv.Load("configs/.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	handlers.RegisterRoutes()

	fmt.Printf("(http://localhost%s)\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		message := "server startup error: " + err.Error()
		if errors.Is(err, syscall.EADDRINUSE) {
			message = "port unavailable: " + port
		}
		log.Fatal(message)
	}

}
