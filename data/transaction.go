package data

import (
	"sync"

	"github.com/Malpizarr/dbproto/dbdata"
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

// InsertWithTransaction performs an insert operation within a transaction context
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

// erforms an update operation within a transaction context
func (t *Table) UpdateWithTransaction(key string, updates Record) error {
	transaction := NewTransaction(t)
	if err := transaction.Start(); err != nil {
		return err // Start returns an error if the transaction cannot be started
	}

	// Tries to udate the record into the table and if it fails it rolls back the transaction
	if err := t.Update(key, updates); err != nil {
		transaction.Rollback()
		return err
	}
	// Commit the transaction if the insert succeeds
	return transaction.Commit()
}

// performs a delete operation within a transaction context.
func (t *Table) DeleteWithTransaction(key string) error {
	transaction := NewTransaction(t)
	if err := transaction.Start(); err != nil {
		return err // Start returns an error if the transaction cannot be started
	}
	// Tries to delete the record into the table and if it fails it rolls back the transaction
	if err := t.Delete(key); err != nil {
		transaction.Rollback()
		return err
	}

	// Commit the transaction if the insert succeeds
	return transaction.Commit()
}
