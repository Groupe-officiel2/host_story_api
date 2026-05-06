package main

import (
	"fmt"
	"log"
	"net/http"

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
	http.ListenAndServe(port, nil)
}
