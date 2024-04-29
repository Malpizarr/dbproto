package data

import (
	"fmt"
	"log"
	"os"
	"path"
	"sync"

	"github.com/Malpizarr/dbproto/dbdata"
	"github.com/Malpizarr/dbproto/utils"

	"google.golang.org/protobuf/proto"
)

type Record map[string]interface{}

type TableReader interface {
	Select(key string) (Record, error)
	Insert(record Record) error
	Update(key string, record Record) error
	Delete(key string) error
}

type Table struct {
	sync.RWMutex
	FilePath   string
	PrimaryKey string
	utils      *utils.Utils
	Indexes    map[string][]*dbdata.Record
	Records    map[string]*dbdata.Record
}

func NewTable(primaryKey, filePath string) *Table {
	dir := path.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}
	log.Printf("Creating table with file path: %s", filePath)
	table := &Table{
		FilePath:   filePath,
		PrimaryKey: primaryKey,
		utils:      utils.NewUtils(),
		Indexes:    make(map[string][]*dbdata.Record),
	}
	if err := table.initializeFileIfNotExists(); err != nil {
		log.Fatalf("Failed to initialize file %s: %v", filePath, err)
	} else {
		log.Printf("File %s initialized successfully.", filePath)
	}
	table.LoadIndexes()
	return table
}

func (t *Table) LoadIndexes() error {
	records, err := t.readRecordsFromFile()
	if err != nil {
		return err
	}

	if t.Indexes == nil {
		t.Indexes = make(map[string][]*dbdata.Record)
	}

	fmt.Println("Loading indexes...")
	for _, record := range records.GetRecords() {
		for key, value := range record.Fields {
			if value != "" {
				t.Indexes[key] = append(t.Indexes[key], record)
				fmt.Printf("Indexed record under key '%s'\n", key)
			}
		}
	}
	fmt.Println("Indexes loaded successfully.")
	return nil
}

func (t *Table) ResetAndLoadIndexes() error {
	t.Indexes = make(map[string][]*dbdata.Record)

	records, err := t.readRecordsFromFile()
	if err != nil {
		return err
	}

	for _, record := range records.GetRecords() {
		for key := range record.Fields {
			t.Indexes[key] = append(t.Indexes[key], record)
		}
	}
	return nil
}

func (t *Table) initializeFileIfNotExists() error {
	if _, err := os.Stat(t.FilePath); os.IsNotExist(err) {
		records := &dbdata.Records{
			Records: make(map[string]*dbdata.Record),
		}
		if err := t.writeRecordsToFile(records); err != nil {
			return fmt.Errorf("failed to write initial data to file: %v", err)
		}
	}
	return nil
}

func (t *Table) Insert(record Record) error {
	t.Lock()
	defer t.Unlock()

	allRecords, err := t.readRecordsFromFile()
	if err != nil {
		return err
	}

	primaryKeyValue := fmt.Sprintf("%v", record[t.PrimaryKey])
	if _, exists := allRecords.Records[primaryKeyValue]; exists {
		return fmt.Errorf("record with primary key %s already exists", primaryKeyValue)
	}

	protoRecord := &dbdata.Record{Fields: make(map[string]string)}
	for key, value := range record {
		val, ok := value.(string)
		if !ok || val == "" {
			return fmt.Errorf("invalid or empty value type for field %s: %v", key, value)
		}
		protoRecord.Fields[key] = val
		if t.Indexes[key] == nil {
			t.Indexes[key] = []*dbdata.Record{}
		}
		t.Indexes[key] = append(t.Indexes[key], protoRecord)
	}

	allRecords.Records[primaryKeyValue] = protoRecord
	return t.writeRecordsToFile(allRecords)
}

func (t *Table) SelectAll() ([]*dbdata.Record, error) {
	t.RLock()
	defer t.RUnlock()
	records, err := t.readRecordsFromFile()
	if err != nil {
		return nil, err
	}
	var allRecords []*dbdata.Record
	for _, record := range records.GetRecords() {
		allRecords = append(allRecords, record)
	}
	return allRecords, nil
}

func (t *Table) Update(key string, updates Record) error {
	t.Lock()
	defer t.Unlock()

	allRecords, err := t.readRecordsFromFile()
	if err != nil {
		return err
	}
	existingRecord, exists := allRecords.Records[key]
	if !exists {
		return fmt.Errorf("Record with key %s not found", key)
	}

	for field, newValue := range updates {
		oldVal, ok := existingRecord.Fields[field]
		if ok {
			newIdxMap := make([]*dbdata.Record, 0)
			for _, r := range t.Indexes[field] {
				if r.Fields[field] != oldVal {
					newIdxMap = append(newIdxMap, r)
				}
			}
			t.Indexes[field] = newIdxMap
		}
		newValStr, ok := newValue.(string)
		if !ok {
			return fmt.Errorf("non-string value for field %s", field)
		}
		existingRecord.Fields[field] = newValStr
		t.Indexes[field] = make([]*dbdata.Record, 0)
		t.Indexes[field] = append(t.Indexes[field], existingRecord)
	}

	return t.writeRecordsToFile(allRecords)
}

func (t *Table) Delete(key string) error {
	t.Lock()
	defer t.Unlock()

	allRecords, err := t.readRecordsFromFile()
	if err != nil {
		return err
	}
	record, exists := allRecords.Records[key]
	if !exists {
		return fmt.Errorf("Record with key %s not found", key)
	}

	for field, value := range record.Fields {
		newIdxMap := make([]*dbdata.Record, 0)
		for _, r := range t.Indexes[field] {
			if r.Fields[field] != value {
				newIdxMap = append(newIdxMap, r)
			}
		}
		t.Indexes[field] = newIdxMap
	}

	delete(allRecords.Records, key)

	return t.writeRecordsToFile(allRecords)
}

func (t *Table) readRecordsFromFile() (*dbdata.Records, error) {
	encryptedData, err := os.ReadFile(t.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &dbdata.Records{Records: make(map[string]*dbdata.Record)}, nil
		}
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	if len(encryptedData) == 0 {
		return &dbdata.Records{Records: make(map[string]*dbdata.Record)}, nil
	}

	decryptedData, err := t.utils.Decrypt(string(encryptedData))
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %v", err)
	}

	var records dbdata.Records
	if err := proto.Unmarshal(decryptedData, &records); err != nil {
		return nil, fmt.Errorf("proto unmarshal failed: %v", err)
	}

	return &records, nil
}

func (t *Table) writeRecordsToFile(records *dbdata.Records) error {
	data, err := proto.Marshal(records)
	if err != nil {
		return fmt.Errorf("error marshaling records: %v", err)
	}
	encryptedData, err := t.utils.Encrypt(data)
	if err != nil {
		return fmt.Errorf("error encrypting data: %v", err)
	}

	file, err := os.OpenFile(t.FilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening file '%s': %v", t.FilePath, err)
	}
	defer file.Close()

	n, err := file.Write([]byte(encryptedData))
	if err != nil {
		return fmt.Errorf("error writing to file '%s': %v", t.FilePath, err)
	}
	log.Printf("Wrote %d bytes to file %s", n, t.FilePath)

	return nil
}
