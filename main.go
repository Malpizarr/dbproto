package main

import (
	"log"
	"net/http"

	"github.com/Malpizarr/dbproto/api"
	"github.com/Malpizarr/dbproto/data"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	server := data.NewServer()

	if err := server.Initialize(); err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	api.SetupRoutes(server)

	log.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe error:", err)
	}
}