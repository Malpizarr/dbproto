package data

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

type Server struct {
	sync.RWMutex                      // Mutex to ensure the server is thread safe
	Databases    map[string]*Database // Map of Databases in the server
}

func NewServer() *Server {
	return &Server{
		Databases: make(map[string]*Database),
	}
}

// Initialize initializes the server by creating the server directory and loading the databases.
func (s *Server) Initialize() error {
	serverDir := getDefaultServerDir()
	if err := os.MkdirAll(serverDir, 0755); err != nil {
		return fmt.Errorf("failed to create or access server directory: %v", err)
	}

	return s.LoadDatabases()
}

// LoadDatabases loads the databases from the server directory.
func (s *Server) LoadDatabases() error {
	dbs, err := os.ReadDir(getDefaultServerDir())
	if err != nil {
		return fmt.Errorf("failed to read server directory: %v", err)
	}

	for _, dbInfo := range dbs {
		if dbInfo.IsDir() {
			dbDir := filepath.Join(getDefaultServerDir(), dbInfo.Name())
			db := NewDatabase(dbInfo.Name())
			if err := db.LoadTables(dbDir); err != nil {
				return err
			}
			s.Databases[dbInfo.Name()] = db
		}
	}
	return nil
}

// getDefaultServerDir returns the default server directory based on the operating system.
func getDefaultServerDir() string {
	var baseDir string
	switch Os := runtime.GOOS; Os {
	case "windows":
		baseDir = os.Getenv("APPDATA")
		if baseDir == "" {
			baseDir = os.Getenv("USERPROFILE")
		}
	case "linux", "darwin":
		baseDir = os.Getenv("HOME")
	default:
		baseDir = "."
	}

	return filepath.Join(baseDir, "DBPROTO", "databases")
}

// getDefaultServerDir returns the default server directory based on the operating system.
func getDefaultBackUpDir() string {
	var baseDir string
	switch Os := runtime.GOOS; Os {
	case "windows":
		baseDir = os.Getenv("APPDATA")
		if baseDir == "" {
			baseDir = os.Getenv("USERPROFILE")
		}
	case "linux", "darwin":
		baseDir = os.Getenv("HOME")
	default:
		baseDir = "."
	}

	return filepath.Join(baseDir, "DBPROTO_backups")
}

// CreateDatabase creates a new database in the server.
func (s *Server) CreateDatabase(name string) error {
	s.Lock()
	defer s.Unlock()
	if _, exists := s.Databases[name]; exists {
		return fmt.Errorf("Database %s already exists", name)
	}
	s.Databases[name] = NewDatabase(name)
	return nil
}

// ListDatabases returns a list of databases in the server.
func (s *Server) ListDatabases() []string {
	s.RLock()
	defer s.RUnlock()
	var databases []string
	for name := range s.Databases {
		databases = append(databases, name)
	}
	return databases
}

// BackupDatabases creates a backup of all databases in the server.
func (s *Server) BackupDatabases() (string, error) {
	s.RLock()
	defer s.RUnlock()

	backupDir := filepath.Join(getDefaultBackUpDir(), "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %v", err)
	}

	backupPath := filepath.Join(backupDir, "backup.zip")
	backupFile, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %v", err)
	}
	defer backupFile.Close()

	zipWriter := zip.NewWriter(backupFile)
	defer zipWriter.Close()

	for dbName := range s.Databases {
		dbDir := filepath.Join(getDefaultServerDir(), dbName)
		err := filepath.Walk(dbDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return nil
			}

			relativePath, err := filepath.Rel(getDefaultServerDir(), path)
			if err != nil {
				return err
			}

			zipFile, err := zipWriter.Create(relativePath)
			if err != nil {
				return err
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(zipFile, file)
			if err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			return "", fmt.Errorf("failed to backup database %s: %v", dbName, err)
		}
	}

	return backupPath, nil
}

// RestoreDatabases restores databases from the latest backup file.
func (s *Server) RestoreDatabases() error {
	s.Lock()
	defer s.Unlock()

	backupPath := filepath.Join(getDefaultBackUpDir(), "backups", "backup.zip")
	backupFile, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %v", err)
	}
	defer backupFile.Close()

	stat, err := backupFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get backup file stat: %v", err)
	}

	zipReader, err := zip.NewReader(backupFile, stat.Size())
	if err != nil {
		return fmt.Errorf("failed to create zip reader: %v", err)
	}

	for _, file := range zipReader.File {
		filePath := filepath.Join(getDefaultServerDir(), file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, 0755)
			continue
		}

		os.MkdirAll(filepath.Dir(filePath), 0755)

		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return fmt.Errorf("failed to open file for writing: %v", err)
		}

		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open zip file for reading: %v", err)
		}

		_, err = io.Copy(outFile, rc)
		if err != nil {
			return fmt.Errorf("failed to write file: %v", err)
		}

		outFile.Close()
		rc.Close()
	}

	return s.LoadDatabases()
}

// ServeHTTP implements the http.Handler interface for the server.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		if r.URL.Path == "/createDatabase" {
			var data struct {
				Name string
			}
			if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := s.CreateDatabase(data.Name); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			fmt.Fprintf(w, "Database '%s' created successfully", data.Name)
			return
		}
	case "GET":
		if r.URL.Path == "/listDatabases" {
			databases := s.ListDatabases()
			resp, err := json.Marshal(databases)
			if err != nil {
				http.Error(w, "Failed to serialize response", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(resp)
			return
		}
	default:
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
	}
}
