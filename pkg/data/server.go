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

// NewServer creates a new Server instance.
// It initializes the Databases field as an empty map where the key is a string representing the database name
// and the value is a pointer to a Database instance.
// It returns a pointer to the newly created Server instance.
func NewServer() *Server {
	return &Server{
		Databases: make(map[string]*Database),
	}
}

// Initialize is a method of the Server struct that initializes the server.
// It creates the server directory and loads the databases.
// The server directory is determined by the getDefaultServerDir function.
// If the server directory does not exist, it is created with read, write, and execute permissions for the user only.
// If there is an error creating the server directory, the error is returned.
// After the server directory is successfully created or if it already exists, the databases are loaded using the LoadDatabases method.
// If there is an error loading the databases, the error is returned.
// If the server directory is successfully created and the databases are successfully loaded, the method returns nil.
func (s *Server) Initialize() error {
	serverDir := getDefaultServerDir()
	if err := os.MkdirAll(serverDir, 0755); err != nil {
		return fmt.Errorf("failed to create or access server directory: %v", err)
	}

	return s.LoadDatabases()
}

// LoadDatabases is a method of the Server struct that loads the databases from the server directory.
// It reads the server directory using the os.ReadDir function and the getDefaultServerDir function.
// If there is an error reading the server directory, the error is returned.
// For each directory in the server directory, it creates a new Database instance with the directory name as the database name.
// It then loads the tables from the database directory using the LoadTables method of the Database struct.
// If there is an error loading the tables, the error is returned.
// If the tables are successfully loaded, the database is added to the Databases field of the Server struct.
// If all databases are successfully loaded, the method returns nil.
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

// BackupDatabases is a method of the Server struct that creates a backup of all databases in the server.
// It returns the path to the backup file and an error if there is one.
//
// The method works as follows:
//  1. It acquires a read lock on the Server struct and defers the unlocking of the lock.
//  2. It creates a backup directory in the default backup directory. The default backup directory is determined by the getDefaultBackUpDir function.
//     If there is an error creating the backup directory, the error is returned.
//  3. It creates a backup file in the backup directory. The backup file is a zip file named "backup.zip".
//     If there is an error creating the backup file, the error is returned.
//  4. It creates a new zip writer for the backup file and defers the closing of the zip writer.
//  5. It iterates over each database in the Databases field of the Server struct.
//     For each database, it walks the database directory and adds each file to the zip file.
//     The database directory is determined by the getDefaultServerDir function and the database name.
//     If there is an error walking the database directory or adding a file to the zip file, the error is returned.
//  6. If all databases are successfully backed up, the method returns the path to the backup file and nil.
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

// RestoreDatabases is a method of the Server struct that restores databases from the latest backup file.
// It acquires a lock on the Server struct and defers the unlocking of the lock.
// It opens the latest backup file in the default backup directory. The default backup directory is determined by the getDefaultBackUpDir function.
// If there is an error opening the backup file, the error is returned.
// It gets the file stat of the backup file. If there is an error getting the file stat, the error is returned.
// It creates a new zip reader for the backup file. If there is an error creating the zip reader, the error is returned.
// It iterates over each file in the zip file.
// For each file, it creates the file path by joining the default server directory and the file name.
// The default server directory is determined by the getDefaultServerDir function.
// If the file is a directory, it creates the directory with read, write, and execute permissions for the user only.
// If the file is a regular file, it creates the file with the same permissions as in the zip file.
// It opens the file for writing. If there is an error opening the file, the error is returned.
// It opens the zip file for reading. If there is an error opening the zip file, the error is returned.
// It copies the contents of the zip file to the file. If there is an error copying the contents, the error is returned.
// It closes the file and the zip file.
// After all files are successfully restored, it loads the databases using the LoadDatabases method of the Server struct.
// If there is an error loading the databases, the error is returned.
// If all databases are successfully loaded, the method returns nil.
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
