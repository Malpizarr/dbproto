package main

import (
	"log"
	"net/http"

	"github.com/Malpizarr/dbproto/data"
)

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

	http.HandleFunc("/listTables", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
			return
		}
		dbName := r.URL.Query().Get("dbName")
		if dbName == "" {
			http.Error(w, "Database name is required", http.StatusBadRequest)
			return
		}
		if db, exists := server.Databases[dbName]; exists {
			db.ListTables(w)
		} else {
			http.Error(w, "Database not found", http.StatusNotFound)
		}
	})

	http.HandleFunc("/", server.ServeHTTP)

	log.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe error:", err)
	}
}
