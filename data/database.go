package data

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type DatabaseReader interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	ListTables(w http.ResponseWriter) []string
}

type Database struct {
	sync.RWMutex
	Tables map[string]*Table
}

func NewDatabase() *Database {
	return &Database{
		Tables: make(map[string]*Table),
	}
}

func (db *Database) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		switch r.URL.Path {
		case "/listTables":
			db.ListTables(w)
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	} else if r.Method == "POST" {
		decoder := json.NewDecoder(r.Body)
		var data struct {
			Action    string
			TableName string
			Record    Record
			Key       string
			Updates   Record
		}
		if err := decoder.Decode(&data); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		db.RLock()
		table, exists := db.Tables[data.TableName]
		db.RUnlock()
		if !exists {
			db.Lock()
			if _, exists := db.Tables[data.TableName]; !exists {
				db.Tables[data.TableName] = NewTable(data.TableName)
			}
			table = db.Tables[data.TableName]
			db.Unlock()
		}
		switch data.Action {
		case "insert":
			err := table.Insert(data.Record)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		case "update":
			err := table.Update(data.Key, data.Updates)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		case "delete":
			err := table.Delete(data.Key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		default:
			http.Error(w, "Invalid action", http.StatusBadRequest)
			return
		}
		fmt.Fprintln(w, "Operation successful")
	} else {
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
	}
}

func (db *Database) ListTables(w http.ResponseWriter) {
	db.RLock()
	defer db.RUnlock()
	tables := make([]string, 0, len(db.Tables))
	for tableName := range db.Tables {
		tables = append(tables, tableName)
	}
	data, err := json.Marshal(tables)
	if err != nil {
		http.Error(w, "Failed to serialize tables", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
