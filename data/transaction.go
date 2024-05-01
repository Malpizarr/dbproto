package data

import (
	"sync"

	"github.com/Malpizarr/dbproto/dbdata"
	"google.golang.org/protobuf/proto"
)

type Transaction struct {
	sync.Mutex
	OriginalRecords map[string]*dbdata.Record
	Table           *Table
}

func NewTransaction(table *Table) *Transaction {
	return &Transaction{
		Table:           table,
		OriginalRecords: make(map[string]*dbdata.Record),
	}
}

func (t *Transaction) Start() error {
	t.Lock()

	records, err := t.Table.readRecordsFromFile()
	if err != nil {
		t.Unlock()
		return err
	}

	for key, record := range records.Records {
		t.OriginalRecords[key] = proto.Clone(record).(*dbdata.Record)
	}
	return nil
}

func (t *Transaction) Commit() error {
	t.Unlock()
	return nil
}

func (t *Transaction) Rollback() error {
	t.Table.Lock()
	defer t.Table.Unlock()
	defer t.Unlock()

	return t.Table.writeRecordsToFile(&dbdata.Records{Records: t.OriginalRecords})
}

func (t *Table) InsertWithTransaction(record Record) error {
	transaction := NewTransaction(t)
	if err := transaction.Start(); err != nil {
		return err
	}

	if err := t.Insert(record); err != nil {
		transaction.Rollback()
		return err
	}

	return transaction.Commit()
}

func (t *Table) UpdateWithTransaction(key string, updates Record) error {
	transaction := NewTransaction(t)
	if err := transaction.Start(); err != nil {
		return err
	}

	if err := t.Update(key, updates); err != nil {
		transaction.Rollback()
		return err
	}

	return transaction.Commit()
}

func (t *Table) DeleteWithTransaction(key string) error {
	transaction := NewTransaction(t)
	if err := transaction.Start(); err != nil {
		return err
	}

	if err := t.Delete(key); err != nil {
		transaction.Rollback()
		return err
	}

	return transaction.Commit()
}
