package data

import (
	"sync"

	"github.com/Malpizarr/dbproto/pkg/dbdata"
	"google.golang.org/protobuf/proto"
)

// This srtuct holds the transaction data for managing the transaction
type Transaction struct {
	sync.Mutex                                // Mutex to ensure the transaction is thread safe
	OriginalRecords map[string]*dbdata.Record // Map to hold the original records
	Table           *Table                    // Table to which the transaction belongs
}

// Creates a new transaction with the table
func NewTransaction(table *Table) *Transaction {
	return &Transaction{
		Table:           table,
		OriginalRecords: make(map[string]*dbdata.Record),
	}
}

// Start begins the transaction by locking the table and duplicating the records for potential rollback
func (t *Transaction) Start() error {
	t.Lock() // Thi is the lock for the transaction to prevent other transactions from happening

	// Read the records from the file
	records, err := t.Table.readRecordsFromFile()
	if err != nil {
		t.Unlock()
		return err
	}
	// This is the loop to clone the records and store them in the original records
	for key, record := range records.Records {
		t.OriginalRecords[key] = proto.Clone(record).(*dbdata.Record)
	}
	return nil
}

// Commit ends the transaction by unlocking it, indicating successful completion of all operations
func (t *Transaction) Commit() error {
	t.Unlock()
	return nil
}

// Rollback ends the transaction by unlocking it and writing the original records back to the file
func (t *Transaction) Rollback() error {
	t.Table.Lock()
	defer t.Table.Unlock()
	defer t.Unlock()

	return t.Table.writeRecordsToFile(&dbdata.Records{Records: t.OriginalRecords})
}

// InsertWithTransaction is a method of the Table struct that performs an insert operation within a transaction context.
// It first creates a new transaction for the table.
// It then starts the transaction by locking the table and duplicating the records for potential rollback.
// If an error occurs while starting the transaction, it returns the error.
// It then tries to insert the record into the table.
// If an error occurs while inserting the record, it rolls back the transaction and returns the error.
// If the insert operation is successful, it commits the transaction and returns nil.
//
// Parameters:
// - record: A Record representing the record to be inserted into the table.
//
// Returns:
// - If the operation is successful, it returns nil.
// - If an error occurs, it returns the error.
func (t *Table) InsertWithTransaction(record Record) error {
	transaction := NewTransaction(t)
	if err := transaction.Start(); err != nil {
		return err // Start returns an error if the transaction cannot be started
	}

	// Tries to insert the record into the table and if it fails it rolls back the transaction
	if err := t.Insert(record); err != nil {
		transaction.Rollback()
		return err
	}
	// Commit the transaction if the insert succeeds
	return transaction.Commit()
}

// UpdateWithTransaction is a method of the Table struct that performs an update operation within a transaction context.
// It first creates a new transaction for the table.
// It then starts the transaction by locking the table and duplicating the records for potential rollback.
// If an error occurs while starting the transaction, it returns the error.
// It then tries to update the record in the table with the given key and updates.
// If an error occurs while updating the record, it rolls back the transaction and returns the error.
// If the update operation is successful, it commits the transaction and returns nil.
//
// Parameters:
// - key: An interface{} representing the key of the record to be updated. The key is converted to a string before the update is performed.
// - updates: A Record representing the fields to be updated in the record. The keys are field names and the values are the new field values.
//
// Returns:
// - If the operation is successful, it returns nil.
// - If an error occurs, it returns the error.
func (t *Table) UpdateWithTransaction(key interface{}, updates Record) error {
	transaction := NewTransaction(t)
	if err := transaction.Start(); err != nil {
		return err // Start returns an error if the transaction cannot be started
	}

	// Tries to update the record in the table and if it fails it rolls back the transaction
	if err := t.Update(key, updates); err != nil {
		transaction.Rollback()
		return err
	}
	// Commit the transaction if the update succeeds
	return transaction.Commit()
}

// DeleteWithTransaction is a method of the Table struct that performs a delete operation within a transaction context.
// It first creates a new transaction for the table.
// It then starts the transaction by locking the table and duplicating the records for potential rollback.
// If an error occurs while starting the transaction, it returns the error.
// It then tries to delete the record from the table with the given key.
// If an error occurs while deleting the record, it rolls back the transaction and returns the error.
// If the delete operation is successful, it commits the transaction and returns nil.
//
// Parameters:
// - key: An interface{} representing the key of the record to be deleted. The key is converted to a string before the deletion is performed.
//
// Returns:
// - If the operation is successful, it returns nil.
// - If an error occurs, it returns the error.
func (t *Table) DeleteWithTransaction(key interface{}) error {
	transaction := NewTransaction(t)
	if err := transaction.Start(); err != nil {
		return err // Start returns an error if the transaction cannot be started
	}
	// Tries to delete the record from the table and if it fails it rolls back the transaction
	if err := t.Delete(key); err != nil {
		transaction.Rollback()
		return err
	}

	// Commit the transaction if the delete operation succeeds
	return transaction.Commit()
}
