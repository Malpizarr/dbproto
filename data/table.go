package data

import (
	"fmt"
	"sync"
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
	Rows       map[string]Record
	PrimaryKey string
}

func NewTable(primaryKey string) *Table {
	return &Table{
		Rows:       make(map[string]Record),
		PrimaryKey: primaryKey,
	}
}

func (t *Table) Select(key string) (Record, error) {
	t.RLock()
	defer t.RUnlock()
	record, exists := t.Rows[key]
	if !exists {
		return nil, fmt.Errorf("Record with key %s not found", key)
	}
	return record, nil
}

func (t *Table) Insert(record Record) error {
	t.Lock()
	defer t.Unlock()
	key := fmt.Sprintf("%v", record[t.PrimaryKey])
	if _, exists := t.Rows[key]; exists {
		return fmt.Errorf("Record with key %s already exists", key)
	}
	t.Rows[key] = record
	return nil
}

func (t *Table) Update(key string, record Record) error {
	t.Lock()
	defer t.Unlock()
	_, exists := t.Rows[key]
	if !exists {
		return fmt.Errorf("Record with key %s not found", key)
	}
	for k, v := range record {
		t.Rows[key][k] = v
	}
	return nil
}

func (t *Table) Delete(key string) error {
	t.Lock()
	defer t.Unlock()
	_, exists := t.Rows[key]
	if !exists {
		return fmt.Errorf("Record with key %s not found", key)
	}
	delete(t.Rows, key)
	return nil
}
