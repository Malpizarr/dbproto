package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Malpizarr/dbproto/data"
)

func CreateDatabaseHandler(server *data.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if err := server.CreateDatabase(payload.Name); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Database '%s' created successfully.", payload.Name)
	}
}

func CreateTableHandler(server *data.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		dbName := r.URL.Query().Get("dbName")
		if dbName == "" {
			http.Error(w, "Database name is required", http.StatusBadRequest)
			return
		}

		var payload struct {
			TableName  string `json:"tableName"`
			PrimaryKey string `json:"primaryKey"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		db, exists := server.Databases[dbName]
		if !exists {
			http.Error(w, "Database not found", http.StatusNotFound)
			return
		}

		if err := db.CreateTable(payload.TableName, payload.PrimaryKey); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Table '%s' created successfully in database '%s'.", payload.TableName, dbName)
	}
}

func ListDatabasesHandler(server *data.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
			return
		}
		databases := server.ListDatabases()
		json.NewEncoder(w).Encode(databases)
	}
}

func TableActionHandler(server *data.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
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

		var payload struct {
			Action    string      `json:"action"`
			TableName string      `json:"tableName"`
			Record    data.Record `json:"record,omitempty"`
			Key       string      `json:"key,omitempty"`
			Updates   data.Record `json:"updates,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		table, exists := db.Tables[payload.TableName]
		if !exists {
			http.Error(w, "Table not found", http.StatusNotFound)
			return
		}

		switch payload.Action {
		case "insert":
			if err := table.Insert(payload.Record); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case "update":
			if err := table.Update(payload.Key, payload.Updates); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case "delete":
			if err := table.Delete(payload.Key); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case "selectAll":
			records, err := table.SelectAll()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(records)
			return
		default:
			http.Error(w, "Invalid action", http.StatusBadRequest)
		}

		fmt.Fprintf(w, "Action '%s' performed successfully on table '%s'.", payload.Action, payload.TableName)
	}
}
