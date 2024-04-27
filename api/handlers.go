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

func JoinTablesHandler(server *data.Server) http.HandlerFunc {
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

		var joinRequest struct {
			Table1   string        `json:"table1"`
			Table2   string        `json:"table2"`
			Key1     string        `json:"key1"`
			Key2     string        `json:"key2"`
			JoinType data.JoinType `json:"joinType"`
		}
		if err := json.NewDecoder(r.Body).Decode(&joinRequest); err != nil {
			fmt.Printf("Error decoding JSON: %v\n", err)
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		db, exists := server.Databases[dbName]
		if !exists {
			http.Error(w, "Database not found", http.StatusNotFound)
			return
		}

		t1, exists1 := db.Tables[joinRequest.Table1]
		t2, exists2 := db.Tables[joinRequest.Table2]
		if !exists1 || !exists2 {
			http.Error(w, "One or both tables not found", http.StatusNotFound)
			return
		}

		results, err := data.JoinTables(t1, t2, joinRequest.Key1, joinRequest.Key2, joinRequest.JoinType)
		if err != nil {
			fmt.Printf("Error joining tables: %v\n", err)
			http.Error(w, "Join operation failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		response, err := json.Marshal(results)
		if err != nil {
			fmt.Printf("Error marshaling response: %v\n", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	}
}
