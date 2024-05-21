package data

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"sync"

	"github.com/Malpizarr/dbproto/pkg/dbdata"
	"github.com/Malpizarr/dbproto/pkg/utils"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type Record map[string]interface{}

// Table is a struct that represents a table in the database.
// It includes a mutex for read-write locking to ensure thread safety during operations.
// FilePath is the path to the file where the table data is stored.
// PrimaryKey is the field name that is used as the primary key for the table.
// utils is a utility object used for various helper functions.
// Indexes is a map where the keys are field names and the values are slices of records that have that field.
// Records is a map where the keys are primary key values and the values are the corresponding records.
type Table struct {
	sync.RWMutex                             // Mutex for read-write locking
	FilePath     string                      // Path to the file where the table data is stored
	PrimaryKey   string                      // Field name used as the primary key for the table
	utils        *utils.Utils                // Utility object used for various helper functions
	Indexes      map[string][]*dbdata.Record // Map of field names to slices of records that have that field
	Records      map[string]*dbdata.Record   // Map of primary key values to the corresponding records
	Cache        map[string]*dbdata.Record   // Cache for recently accessed records
	metrics      *Metrics                    // Metrics for monitoring
}

// NewTable is a constructor function for the Table struct.
// It takes a primary key and a file path as arguments and returns a pointer to a new Table instance.
//
// The function first gets the directory from the file path and checks if it exists.
// If the directory does not exist, it creates it with the appropriate permissions.
// It then creates a new Table instance, setting the FilePath, PrimaryKey, utils, Records, and Indexes fields.
// It calls the initializeFileIfNotExists method to ensure that the file where the table data is stored exists.
// If the file does not exist, it is created and initialized with an empty Records map.
// If an error occurs during this operation, the function logs the error and exits.
// It then calls the LoadIndexes method to load the indexes from the file into the Indexes map.
// If an error occurs during this operation, the function logs the error and exits.
// Finally, it returns the new Table instance.
//
// Parameters:
// - primaryKey: A string representing the field name to be used as the primary key for the table.
// - filePath: A string representing the path to the file where the table data is stored.
//
// Returns:
// - A pointer to a new Table instance.
func NewTable(primaryKey, filePath string) *Table {
	dir := path.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}
	table := &Table{
		FilePath:   filePath,
		PrimaryKey: primaryKey,
		utils:      utils.NewUtils(),
		Records:    make(map[string]*dbdata.Record),
		Indexes:    make(map[string][]*dbdata.Record),
		Cache:      make(map[string]*dbdata.Record),
		metrics:    NewMetrics(),
	}
	if err := table.initializeFileIfNotExists(); err != nil {
		log.Fatalf("Failed to initialize file %s: %v", filePath, err)
	}
	err := table.LoadIndexes()
	if err != nil {
		log.Fatalf("Failed to load indexes: %v", err)
	}
	return table
}

// LoadIndexes loads the indexes from the file
func (t *Table) LoadIndexes() error {
	records, err := t.readRecordsFromFile()
	if err != nil {
		return err
	}

	if t.Indexes == nil {
		t.Indexes = make(map[string][]*dbdata.Record)
	}

	for _, record := range records.GetRecords() {
		for key, value := range record.Fields {
			if value != nil && value.GetStringValue() != "" {
				t.Indexes[key] = append(t.Indexes[key], record)
			}
		}
	}
	return nil
}

// ResetAndLoadIndexes resets the indexes and reloads them from the file
func (t *Table) ResetAndLoadIndexes() error {
	t.Lock()
	defer t.Unlock()

	t.Indexes = make(map[string][]*dbdata.Record)

	records, err := t.readRecordsFromFile()
	if err != nil {
		return fmt.Errorf("failed to read records from file: %v", err)
	}

	for _, record := range records.GetRecords() {
		for key, value := range record.Fields {
			if value != nil && value.GetStringValue() != "" {
				t.Indexes[key] = append(t.Indexes[key], record)
			}
		}
	}

	return nil
}

// initializeFileIfNotExists is a method of the Table struct that initializes the file if it doesn't exist.
// It first checks if the file at the specified file path exists.
// If the file does not exist, it creates a new dbdata.Records struct, initializes its Records map, and writes this initial data to the file.
// If an error occurs while writing to the file, it returns the error.
// If the file already exists or if the initial data is successfully written to the file, it returns nil.
//
// Returns:
// - If the operation is successful, it returns nil.
// - If an error occurs while writing to the file, it returns the error.
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

// Insert is a method of the Table struct that inserts a new record into the table.
// It locks the table for writing, ensuring that no other goroutines can modify the table while the insertion is happening.
// It first checks if the Indexes map is initialized, and if not, it initializes it.
// It then reads all existing records from the file where the table data is stored.
// If the primary key of the new record already exists in the table, it returns an error.
// It then creates a new proto Record from the input record, converting each field value to a proto Value.
// For each field in the new record, it adds the new record to the index for that field.
// If the index for a field does not exist, it initializes it before adding the new record.
// It then adds the new record to the main records map and writes the updated records back to the file.
// If any error occurs during these operations, it returns the error.
//
// Parameters:
// - record: A map representing the record to be inserted. The keys are field names and the values are the field values.
//
// Returns:
// - If the operation is successful, it returns nil.
// - If an error occurs, it returns the error.

func (t *Table) Insert(record Record) error {
	primaryKeyValue, ok := record[t.PrimaryKey]
	if !ok {
		return fmt.Errorf("primary key '%s' not found in record", t.PrimaryKey)
	}

	primaryKeyString := fmt.Sprintf("%v", primaryKeyValue)
	if primaryKeyString == "<nil>" || primaryKeyString == "" {
		return fmt.Errorf("primary key '%s' is nil or empty", t.PrimaryKey)
	}

	protoRecord := &dbdata.Record{Fields: make(map[string]*structpb.Value)}
	for key, value := range record {
		protoValue, err := structpb.NewValue(value)
		if err != nil {
			return fmt.Errorf("invalid value type for field '%s': %v", key, err)
		}
		protoRecord.Fields[key] = protoValue
	}

	t.Lock()
	defer t.Unlock()

	if t.Indexes == nil {
		t.Indexes = make(map[string][]*dbdata.Record)
	}

	allRecords, err := t.readRecordsFromFile()
	if err != nil {
		return err
	}

	if _, exists := allRecords.Records[primaryKeyString]; exists {
		return fmt.Errorf("record with primary key '%s' already exists", primaryKeyString)
	}

	for key := range protoRecord.Fields {
		if _, exists := t.Indexes[key]; !exists {
			t.Indexes[key] = []*dbdata.Record{}
		}
		t.Indexes[key] = append(t.Indexes[key], protoRecord)
	}

	allRecords.Records[primaryKeyString] = protoRecord
	t.Cache[primaryKeyString] = protoRecord

	t.metrics.IncrementInsertCount()
	return t.writeRecordsToFile(allRecords)
}

// InsertMany is a method of the Table struct that inserts multiple new records into the table.
// It divides the records into batches and inserts each batch separately to optimize performance.
// If an error occurs while inserting a batch, it is added to a slice of errors.
// After all batches have been processed, it returns the slice of errors.
//
// Parameters:
// - records: A slice of maps representing the records to be inserted. The keys are field names and the values are the field values.
//
// Returns:
// - A slice of errors for records that failed to insert. If all records are inserted successfully, the slice is empty.
func (t *Table) InsertMany(records []Record) []error {
	batchSize := 1000
	numRecords := len(records)
	errors := make([]error, 0)

	for i := 0; i < numRecords; i += batchSize {
		end := i + batchSize
		if end > numRecords {
			end = numRecords
		}

		batch := records[i:end]
		batchErrors := t.insertBatch(batch)
		if batchErrors != nil {
			errors = append(errors, batchErrors...)
		}
	}

	return errors
}

// insertBatch is a helper method of the Table struct that inserts a batch of records into the table.
// It first validates the records, checking for missing primary keys, nil or empty primary keys, and duplicate primary keys within the batch.
// It then locks the table for writing, ensuring that no other goroutines can modify the table while the insertion is happening.
// It reads all existing records from the file where the table data is stored.
// For each record in the batch, if the primary key already exists in the table, it returns an error for that record but continues with the rest.
// It then updates the indexes and writes the updated records back to the file.
// If any error occurs during these operations, it returns the error.
//
// Parameters:
// - records: A slice of maps representing the records to be inserted. The keys are field names and the values are the field values.
//
// Returns:
// - A slice of errors for records that failed to insert. If all records are inserted successfully, the slice is empty.
func (t *Table) insertBatch(records []Record) []error {
	protoRecords := make(map[string]*dbdata.Record)
	errors := make([]error, 0)

	for _, record := range records {
		primaryKeyValue, ok := record[t.PrimaryKey]
		if !ok {
			errors = append(errors, fmt.Errorf("primary key '%s' not found in record", t.PrimaryKey))
			continue
		}

		primaryKeyString := fmt.Sprintf("%v", primaryKeyValue)
		if primaryKeyString == "<nil>" || primaryKeyString == "" {
			errors = append(errors, fmt.Errorf("primary key '%s' is nil or empty", t.PrimaryKey))
			continue
		}

		if _, exists := protoRecords[primaryKeyString]; exists {
			errors = append(errors, fmt.Errorf("duplicate primary key '%s' in input records", primaryKeyString))
			continue
		}

		protoRecord := &dbdata.Record{Fields: make(map[string]*structpb.Value)}
		for key, value := range record {
			protoValue, err := structpb.NewValue(value)
			if err != nil {
				errors = append(errors, fmt.Errorf("invalid value type for field '%s': %v", key, err))
				continue
			}
			protoRecord.Fields[key] = protoValue
		}

		protoRecords[primaryKeyString] = protoRecord
	}

	t.Lock()
	defer t.Unlock()

	if t.Indexes == nil {
		t.Indexes = make(map[string][]*dbdata.Record)
	}

	allRecords, err := t.readRecordsFromFile()
	if err != nil {
		return append(errors, err)
	}

	for primaryKeyString, protoRecord := range protoRecords {
		if _, exists := allRecords.Records[primaryKeyString]; exists {
			errors = append(errors, fmt.Errorf("record with primary key '%s' already exists", primaryKeyString))
			continue
		}

		for key := range protoRecord.Fields {
			if _, exists := t.Indexes[key]; !exists {
				t.Indexes[key] = []*dbdata.Record{}
			}
			t.Indexes[key] = append(t.Indexes[key], protoRecord)
		}

		allRecords.Records[primaryKeyString] = protoRecord
		t.Cache[primaryKeyString] = protoRecord
		t.metrics.IncrementInsertCount()
	}

	if err := t.writeRecordsToFile(allRecords); err != nil {
		return append(errors, err)
	}

	return errors
}

// SelectAll is a method of the Table struct that selects all records from the table.
// It locks the table for reading, ensuring that no other goroutines can modify the table while the selection is happening.
// It then reads all existing records from the file where the table data is stored.
// It iterates over the records, appending each one to a slice of records.
// If any error occurs during these operations, it returns the error and a nil slice.
// If the operation is successful, it returns the slice of all records and a nil error.
//
// Returns:
// - A slice of pointers to dbdata.Record instances representing all records in the table.
// - If an error occurs, it returns the error and a nil slice.
// - If the operation is successful, it returns the slice of all records and a nil error.
func (t *Table) SelectAll() ([]*dbdata.Record, error) {
	t.RLock()
	defer t.RUnlock()

	index, exists := t.Indexes[t.PrimaryKey]
	if !exists {

		records, err := t.readRecordsFromFile()
		if err != nil {
			return nil, err
		}
		var allRecords []*dbdata.Record
		for _, record := range records.GetRecords() {
			allRecords = append(allRecords, record)
		}
		t.metrics.IncrementQueryCount()
		return allRecords, nil
	}

	var allRecords []*dbdata.Record
	for _, record := range index {
		allRecords = append(allRecords, record)
	}
	t.metrics.IncrementQueryCount()
	return allRecords, nil
}

// Equal checks if two structpb.Value are equal
func Equal(value1, value2 *structpb.Value) bool {
	if value1.GetKind() == nil || value2.GetKind() == nil {
		return false
	}

	switch value1.GetKind().(type) {
	case *structpb.Value_NumberValue:
		return value1.GetNumberValue() == value2.GetNumberValue()
	case *structpb.Value_StringValue:
		return value1.GetStringValue() == value2.GetStringValue()
	case *structpb.Value_BoolValue:
		return value1.GetBoolValue() == value2.GetBoolValue()
	case *structpb.Value_StructValue:
		return false
	case *structpb.Value_ListValue:

		return false
	default:
		return false
	}
}

// SelectWithFilter is a method of the Table struct that selects records from the table based on the given filters.
// It locks the table for reading, ensuring that no other goroutines can modify the table while the selection is happening.
// It then reads all existing records from the file where the table data is stored.
// It iterates over the records, checking each one against the filters.
// For each record, it iterates over the filters. For each filter, it converts the filter value to a proto Value.
// If an error occurs during this conversion, it returns the error and a nil slice.
// It then checks if the field specified by the filter exists in the record and if the value of the field in the record is equal to the filter value.
// If the field does not exist in the record or if the values are not equal, it skips to the next record.
// If all filters match for a record, it appends the record to a slice of matched records.
// If the operation is successful, it returns the slice of matched records and a nil error.
//
// Parameters:
// - filters: A map where the keys are field names and the values are the filter values. Only records where the field values match the filter values are selected.
//
// Returns:
// - A slice of pointers to dbdata.Record instances representing the records that match the filters.
// - If an error occurs, it returns the error and a nil slice.
// - If the operation is successful, it returns the slice of matched records and a nil error.
func (t *Table) SelectWithFilter(filters map[string]interface{}) ([]*dbdata.Record, error) {
	t.RLock()
	defer t.RUnlock()

	allRecords, err := t.readRecordsFromFile()
	if err != nil {
		return nil, err
	}

	var matchedRecords []*dbdata.Record

RecordsLoop:
	for _, record := range allRecords.GetRecords() {
		for field, filterValue := range filters {
			protoValue, err := structpb.NewValue(filterValue)
			if err != nil {
				return nil, fmt.Errorf("error converting filter value for field %s: %v", field, err)
			}
			value, exists := record.Fields[field]
			if !exists || !Equal(value, protoValue) {
				continue RecordsLoop
			}
		}

		matchedRecords = append(matchedRecords, record)
	}

	return matchedRecords, nil
}

// Select is a method of the Table struct that selects a record from the table based on the given key.
// It locks the table for reading, ensuring that no other goroutines can modify the table while the selection is happening.
// It then reads all existing records from the file where the table data is stored.
// It converts the key to a string and checks if a record with that key exists in the table.
// If a record with that key does not exist, it returns an error and a nil record.
// If a record with that key exists, it returns the record and a nil error.
//
// Parameters:
// - key: An interface{} representing the key of the record to be selected. The key is converted to a string before the selection is performed.
//
// Returns:
// - A pointer to a dbdata.Record instance representing the record with the given key.
// - If a record with the given key does not exist, it returns an error and a nil record.
// - If an error occurs while reading the records from the file, it returns the error and a nil record.
// - If the operation is successful, it returns the record with the given key and a nil error.
func (t *Table) Select(key interface{}) (*dbdata.Record, error) {
	t.RLock()
	defer t.RUnlock()

	keyStr := fmt.Sprintf("%v", key)

	if record, exists := t.Cache[keyStr]; exists {
		t.metrics.IncrementCacheHits()
		return record, nil
	}

	records, err := t.readRecordsFromFile()
	if err != nil {
		return nil, err
	}

	record, exists := records.Records[keyStr]
	if !exists {
		return nil, fmt.Errorf("record with key %s not found", keyStr)
	}

	t.Cache[keyStr] = record
	t.metrics.IncrementCacheMisses()
	t.metrics.IncrementQueryCount()
	return record, nil
}

// Update is a method of the Table struct that updates a record in the table based on the given key.
// It locks the table for writing, ensuring that no other goroutines can modify the table while the update is happening.
// It first reads all existing records from the file where the table data is stored.
// If the primary key of the record to be updated does not exist in the table, it returns an error.
// It then iterates over the fields in the updates map, updating each field in the existing record.
// For each field, it checks if the field exists in the existing record.
// If the field exists, it removes the existing record from the index for that field.
// It then converts the new field value to a proto Value and updates the field in the existing record.
// If an error occurs during this conversion, it returns the error.
// It then adds the updated record to the index for the field.
// If the index for the field does not exist, it initializes it before adding the updated record.
// It then writes the updated records back to the file.
// If any error occurs during these operations, it returns the error.
//
// Parameters:
// - key: An interface{} representing the key of the record to be updated. The key is converted to a string before the update is performed.
// - updates: A map representing the fields to be updated in the record. The keys are field names and the values are the new field values.
//
// Returns:
// - If the operation is successful, it returns nil.
// - If an error occurs, it returns the error.
func (t *Table) Update(key interface{}, updates Record) error {
	t.Lock()
	defer t.Unlock()

	keyStr := fmt.Sprintf("%v", key)
	allRecords, err := t.readRecordsFromFile()
	if err != nil {
		return err
	}
	existingRecord, exists := allRecords.Records[keyStr]
	if !exists {
		return fmt.Errorf("record with key %s not found", keyStr)
	}

	for field, newValue := range updates {
		oldVal := existingRecord.Fields[field]
		if oldVal != nil {
			newIdxMap := make([]*dbdata.Record, 0)
			for _, r := range t.Indexes[field] {
				if r.Fields[field] != oldVal {
					newIdxMap = append(newIdxMap, r)
				}
			}
			t.Indexes[field] = newIdxMap
		}
		newVal, err := structpb.NewValue(newValue)
		if err != nil {
			return fmt.Errorf("error converting newValue for field %s: %v", field, err)
		}
		existingRecord.Fields[field] = newVal
		t.Indexes[field] = append(t.Indexes[field], existingRecord)
	}

	t.Cache[keyStr] = existingRecord

	t.metrics.IncrementUpdateCount()
	return t.writeRecordsToFile(allRecords)
}

// UpdateMany is a method of the Table struct that updates multiple records in the table based on the given keys and updates.
// It locks the table for writing, ensuring that no other goroutines can modify the table while the updates are happening.
// It first reads all existing records from the file where the table data is stored.
// For each key, if the primary key of the record to be updated does not exist in the table, it returns an error for that key but continues with the rest.
// It then iterates over the fields in the updates map, updating each field in the existing record.
// For each field, it checks if the field exists in the existing record.
// If the field exists, it removes the existing record from the index for that field.
// It then converts the new field value to a proto Value and updates the field in the existing record.
// If an error occurs during this conversion, it returns an error for that record but continues with the rest.
// It then adds the updated record to the index for the field.
// If the index for the field does not exist, it initializes it before adding the updated record.
// It then writes the updated records back to the file.
// If any error occurs during these operations, it returns an error for that record but continues with the rest.
//
// Parameters:
// - updates: A map where the keys are the primary key values and the values are the updates to be applied to the records. Each value is a map representing the fields to be updated.
//
// Returns:
// - A slice of errors for records that failed to update. If all records are updated successfully, the slice is empty.
func (t *Table) UpdateMany(updates map[string]Record) []error {
	t.Lock()
	defer t.Unlock()

	allRecords, err := t.readRecordsFromFile()
	if err != nil {
		return []error{fmt.Errorf("failed to read records from file: %w", err)}
	}

	var errors []error

	for keyStr, updateFields := range updates {
		existingRecord, exists := allRecords.Records[keyStr]
		if !exists {
			errors = append(errors, fmt.Errorf("record with key %s not found", keyStr))
			continue
		}

		for field, newValue := range updateFields {
			oldVal := existingRecord.Fields[field]
			if oldVal != nil {
				newIdxMap := make([]*dbdata.Record, 0)
				for _, r := range t.Indexes[field] {
					if r.Fields[field] != oldVal {
						newIdxMap = append(newIdxMap, r)
					}
				}
				t.Indexes[field] = newIdxMap
			}
			newVal, err := structpb.NewValue(newValue)
			if err != nil {
				errors = append(errors, fmt.Errorf("error converting newValue for field %s in record with key %s: %v", field, keyStr, err))
				continue
			}
			existingRecord.Fields[field] = newVal
			t.Indexes[field] = append(t.Indexes[field], existingRecord)
		}

		t.Cache[keyStr] = existingRecord
		t.metrics.IncrementUpdateCount()
	}

	if writeErr := t.writeRecordsToFile(allRecords); writeErr != nil {
		return append(errors, fmt.Errorf("failed to write records to file: %w", writeErr))
	}

	return errors
}

// Delete is a method of the Table struct that deletes a record from the table based on the given key.
// It locks the table for writing, ensuring that no other goroutines can modify the table while the deletion is happening.
// It first reads all existing records from the file where the table data is stored.
// If the primary key of the record to be deleted does not exist in the table, it returns an error.
// It then removes the record from the main records map.
// It also updates the indexes. For each field in the record, it removes the record from the index for that field.
// If the index for a field becomes empty after the removal of the record, it deletes the index.
// It then writes the updated records back to the file.
// If any error occurs during these operations, it returns the error.
//
// Parameters:
// - key: An interface{} representing the key of the record to be deleted. The key is converted to a string before the deletion is performed.
//
// Returns:
// - If the operation is successful, it returns nil.
// - If an error occurs, it returns the error.
func (t *Table) Delete(key interface{}) error {
	t.Lock()
	defer t.Unlock()

	keyStr := fmt.Sprintf("%v", key)

	allRecords, err := t.readRecordsFromFile()
	if err != nil {
		return err
	}

	record, exists := allRecords.Records[keyStr]
	if !exists {
		return fmt.Errorf("record with key %s not found", keyStr)
	}

	delete(allRecords.Records, keyStr)
	delete(t.Cache, keyStr)

	for field := range record.Fields {
		idxSlice := t.Indexes[field]
		for i, rec := range idxSlice {
			if recKeyValue, ok := rec.Fields[t.PrimaryKey]; ok && recKeyValue.GetStringValue() == keyStr {
				t.Indexes[field] = append(idxSlice[:i], idxSlice[i+1:]...)
				break
			}
		}
		if len(t.Indexes[field]) == 0 {
			delete(t.Indexes, field)
		}
	}

	t.metrics.IncrementDeleteCount()
	return t.writeRecordsToFile(allRecords)
}

// DeleteMany is a method of the Table struct that deletes multiple records from the table based on the given keys.
// It locks the table for writing, ensuring that no other goroutines can modify the table while the deletion is happening.
// It first reads all existing records from the file where the table data is stored.
// For each key, if the primary key of the record to be deleted does not exist in the table, it returns an error for that key but continues with the rest.
// It then removes the record from the main records map.
// It also updates the indexes. For each field in the record, it removes the record from the index for that field.
// If the index for a field becomes empty after the removal of the record, it deletes the index.
// It then writes the updated records back to the file.
// If any error occurs during these operations, it returns the error.
//
// Parameters:
// - keys: A slice of interface{} representing the keys of the records to be deleted. The keys are converted to strings before the deletion is performed.
//
// Returns:
// - A slice of errors for keys that failed to delete. If all records are deleted successfully, the slice is empty.
func (t *Table) DeleteMany(keys []interface{}) []error {
	t.Lock()
	defer t.Unlock()

	allRecords, err := t.readRecordsFromFile()
	if err != nil {
		return []error{fmt.Errorf("failed to read records from file: %w", err)}
	}

	var errors []error

	for _, key := range keys {
		keyStr := fmt.Sprintf("%v", key)
		record, exists := allRecords.Records[keyStr]
		if !exists {
			errors = append(errors, fmt.Errorf("record with key %s not found", keyStr))
			continue
		}

		delete(allRecords.Records, keyStr)
		delete(t.Cache, keyStr)

		for field := range record.Fields {
			idxSlice := t.Indexes[field]
			for i, rec := range idxSlice {
				if recKeyValue, ok := rec.Fields[t.PrimaryKey]; ok && recKeyValue.GetStringValue() == keyStr {
					t.Indexes[field] = append(idxSlice[:i], idxSlice[i+1:]...)
					break
				}
			}
			if len(t.Indexes[field]) == 0 {
				delete(t.Indexes, field)
			}
		}

		t.metrics.IncrementDeleteCount()
	}

	if writeErr := t.writeRecordsToFile(allRecords); writeErr != nil {
		errors = append(errors, fmt.Errorf("failed to write records to file: %w", writeErr))
	}

	return errors
}

// readRecordsFromFile reads the records from the file
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

	if records.Records == nil {
		records.Records = make(map[string]*dbdata.Record)
	}

	return &records, nil
}

// writeRecordsToFile writes the records to the file
func (t *Table) writeRecordsToFile(records *dbdata.Records) error {
	data, err := proto.Marshal(records)
	if err != nil {
		return fmt.Errorf("error marshaling records: %v", err)
	}
	encryptedData, err := t.utils.Encrypt(data)
	if err != nil {
		return fmt.Errorf("error encrypting data: %v", err)
	}

	// Use batch writing with buffer
	file, err := os.OpenFile(t.FilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening file '%s': %v", t.FilePath, err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.Write([]byte(encryptedData))
	if err != nil {
		return fmt.Errorf("error writing to file '%s': %v", t.FilePath, err)
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("error flushing writer: %v", err)
	}

	t.Records = records.Records

	return nil
}
