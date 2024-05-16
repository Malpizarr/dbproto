package data

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type DatabaseReader interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	ListTables(w http.ResponseWriter) []string
}

type Database struct {
	sync.RWMutex                   // Mutex to ensure the database is thread safe
	Name         string            // Name of the database
	Tables       map[string]*Table // Map of Tables in the database
}

func NewDatabase(name string) *Database {
	return &Database{
		Name:   name,
		Tables: make(map[string]*Table),
	}
}

func ValidFilename(name string) bool {
	validName := regexp.MustCompile(`^[a-zA-Z0-9-_]+$`).MatchString
	return validName(name)
}

// CreateTable is a method of the Database struct that creates a new table in the database.
// It takes a table name and a primary key as arguments.
// The table name and primary key must match the regex `^[a-zA-Z0-9-_]+$`, meaning they can only contain
// alphanumeric characters, hyphens, and underscores. They cannot contain spaces, punctuation (except for hyphens and underscores),
// or special characters.
// It first checks if the table name and the primary key are valid using the ValidFilename function.
// If either the table name or the primary key is not valid, it returns an error.
// It then acquires a lock on the Database struct and defers the unlocking of the lock.
// It checks if a table with the same name already exists in the database.
// If a table with the same name already exists, it returns an error.
// It then creates the database directory if it does not exist.
// If there is an error creating the database directory, the error is returned.
// It creates a new Table instance with the primary key and the file path of the table.
// It adds the table to the Tables field of the Database struct.
// It then saves the primary key in a metadata file.
// If there is an error serializing the metadata or writing the metadata file, the error is returned.
// It then creates the initial file for the table.
// If there is an error creating the initial file, the error is returned.
// If the table is successfully created, the method returns nil.
func (db *Database) CreateTable(tableName, primaryKey string) error {
	if !ValidFilename(tableName) {
		return fmt.Errorf("invalid table name: %s", tableName)
	}
	if !ValidFilename(primaryKey) {
		return fmt.Errorf("invalid primary key: %s", primaryKey)
	}
	db.Lock()
	defer db.Unlock()
	if _, exists := db.Tables[tableName]; exists {
		return fmt.Errorf("table %s already exists", tableName)
	}

	serverDir := getDefaultServerDir()
	dbDir := filepath.Join(serverDir, db.Name)
	filePath := filepath.Join(dbDir, tableName+".dat")
	metaFilePath := filepath.Join(dbDir, tableName+".meta")

	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %v", err)
	}

	table := NewTable(primaryKey, filePath)
	db.Tables[tableName] = table

	// Save the primary key in a metadata file
	metaData := map[string]string{"PrimaryKey": primaryKey}
	metaDataBytes, err := json.Marshal(metaData)
	if err != nil {
		return fmt.Errorf("failed to serialize metadata: %v", err)
	}
	if err := os.WriteFile(metaFilePath, metaDataBytes, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %v", err)
	}

	if _, err := os.Create(filePath); err != nil {
		return fmt.Errorf("failed to create initial file for table '%s': %v", tableName, err)
	}

	return nil
}

// LoadTables loads the tables from the database directory.
func (db *Database) LoadTables(dbDir string) error {
	files, err := os.ReadDir(dbDir)
	if err != nil {
		return fmt.Errorf("failed to read database directory: %v", err)
	}

	for _, fileInfo := range files {
		if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".dat") {
			tableName := strings.TrimSuffix(fileInfo.Name(), ".dat")
			tablePath := filepath.Join(dbDir, fileInfo.Name())

			// Load the primary key from the metadata file
			metaFilePath := filepath.Join(dbDir, tableName+".meta")
			metaDataBytes, err := os.ReadFile(metaFilePath)
			if err != nil {
				return fmt.Errorf("failed to read metadata file for table %s: %v", tableName, err)
			}
			var metaData map[string]string
			if err := json.Unmarshal(metaDataBytes, &metaData); err != nil {
				return fmt.Errorf("failed to deserialize metadata for table %s: %v", tableName, err)
			}
			primaryKey := metaData["PrimaryKey"]

			table := NewTable(primaryKey, tablePath)
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

// ListTables returns a list of tables in the database
func (db *Database) ListTables() ([]string, error) {
	db.RLock()
	defer db.RUnlock()

	tables := make([]string, 0, len(db.Tables))
	for tableName := range db.Tables {
		tables = append(tables, tableName)
	}

	return tables, nil
}
