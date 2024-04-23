package data

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
)

type DatabaseReader interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	ListTables(w http.ResponseWriter) []string
}

type Database struct {
	sync.RWMutex
	Name   string
	Tables map[string]*Table
}

func NewDatabase(name string) *Database {
	return &Database{
		Name:   name,
		Tables: make(map[string]*Table),
	}
}

func (db *Database) CreateTable(tableName, primaryKey string) error {
	db.Lock()
	defer db.Unlock()
	if _, exists := db.Tables[tableName]; exists {
		return fmt.Errorf("table %s already exists", tableName)
	}
	filePath := fmt.Sprintf("./%s/%s.dat", db.Name, tableName)
	table := NewTable(primaryKey, filePath)
	db.Tables[tableName] = table

	if _, err := os.Create(filePath); err != nil {
		return fmt.Errorf("failed to create initial file for table '%s': %v", tableName, err)
	}

	return nil
}

func (db *Database) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Table not found", http.StatusNotFound)
		return
	}

	switch data.Action {
	case "insert":
		err := table.Insert(data.Record)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	case "selectAll":

		records, err := table.SelectAll()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = json.NewEncoder(w).Encode(records)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
