package main

import (
	"log"

	"github.com/joho/godotenv"
)

func main() {
	envPath := "../../.env"
	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// server := data.NewServer()

	// if err := server.Initialize(); err != nil {
	// 	log.Fatalf("Failed to initialize server: %v", err)
	// }

	// api.SetupRoutes(server)

	// log.Println("Server starting on port 8080...")
	// if err := http.ListenAndServe(":8080", nil); err != nil {
	// 	log.Fatal("ListenAndServe error:", err)
	// }
}
