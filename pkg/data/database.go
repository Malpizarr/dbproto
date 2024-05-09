package data

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

	serverDir := getDefaultServerDir()
	dbDir := filepath.Join(serverDir, db.Name)
	filePath := filepath.Join(dbDir, tableName+".dat")

	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %v", err)
	}

	table := NewTable(primaryKey, filePath)
	db.Tables[tableName] = table

	if _, err := os.Create(filePath); err != nil {
		return fmt.Errorf("failed to create initial file for table '%s': %v", tableName, err)
	}

	return nil
}

func (db *Database) LoadTables(dbDir string) error {
	files, err := os.ReadDir(dbDir)
	if err != nil {
		return fmt.Errorf("failed to read database directory: %v", err)
	}

	for _, fileInfo := range files {
		if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".dat") {
			tableName := strings.TrimSuffix(fileInfo.Name(), ".dat")
			tablePath := filepath.Join(dbDir, fileInfo.Name())
			table := NewTable("", tablePath)
			records, err := table.readRecordsFromFile()
			if err != nil {
				return fmt.Errorf("failed to load table %s: %v", tableName, err)
			}

			table.Records = records.Records
			db.Tables[tableName] = table
		}
	}
	return nil
}

func (db *Database) ListTables() ([]string, error) {
	db.RLock()
	defer db.RUnlock()

	tables := make([]string, 0, len(db.Tables))
	for tableName := range db.Tables {
		tables = append(tables, tableName)
	}

	return tables, nil
}
