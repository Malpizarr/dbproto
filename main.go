package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Malpizarr/dbproto/data"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func main() {
	server := data.NewServer()

	http.HandleFunc("/createDatabase", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}
		server.ServeHTTP(w, r)
	})

	http.HandleFunc("/listDatabases", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
			return
		}
		server.ServeHTTP(w, r)
	})

	http.HandleFunc("/createTable", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}
		decoder := json.NewDecoder(r.Body)
		var data struct {
			TableName  string `json:"TableName"`
			PrimaryKey string `json:"PrimaryKey"`
		}
		if err := decoder.Decode(&data); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		dbName := r.URL.Query().Get("dbName")
		if dbName == "" {
			http.Error(w, "Database name is required", http.StatusBadRequest)
			return
		}
		db, exists := server.Databases[dbName]
		if !exists {
			http.Error(w, "Database not found", http.StatusNotFound)
			return
		}
		if err := db.CreateTable(data.TableName, data.PrimaryKey); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Fprintf(w, "Table '%s' created successfully", data.TableName)
	})

	http.HandleFunc("/tableAction", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}
		dbName := r.URL.Query().Get("dbName")
		if db, exists := server.Databases[dbName]; exists {
			db.ServeHTTP(w, r)
		} else {
			http.Error(w, "Database not found", http.StatusNotFound)
		}
	})

	log.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe error:", err)
	}
}
