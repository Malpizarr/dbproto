package data

import (
	"encoding/json"
	"sync"
	"time"
)

// Metrics is a structure that holds various counts and timestamps related to database operations and cache usage.
type Metrics struct {
	sync.RWMutex
	InsertCount int       // The number of insert operations performed.
	UpdateCount int       // The number of update operations performed.
	DeleteCount int       // The number of delete operations performed.
	QueryCount  int       // The number of query operations performed.
	CacheHits   int       // The number of successful cache retrievals.
	CacheMisses int       // The number of unsuccessful cache retrievals.
	LastInsert  time.Time // The timestamp of the last insert operation.
	LastUpdate  time.Time // The timestamp of the last update operation.
	LastDelete  time.Time // The timestamp of the last delete operation.
	LastQuery   time.Time // The timestamp of the last query operation.
}

// NewMetrics creates and returns a new Metrics structure.
func NewMetrics() *Metrics {
	return &Metrics{}
}

// IncrementInsertCount increases the count of insert operations and updates the timestamp of the last insert operation.
func (m *Metrics) IncrementInsertCount() {
	m.Lock()
	m.InsertCount++
	m.LastInsert = time.Now()
	m.Unlock()
}

// IncrementUpdateCount increases the count of update operations and updates the timestamp of the last update operation.
func (m *Metrics) IncrementUpdateCount() {
	m.Lock()
	m.UpdateCount++
	m.LastUpdate = time.Now()
	m.Unlock()
}

// IncrementDeleteCount increases the count of delete operations and updates the timestamp of the last delete operation.
func (m *Metrics) IncrementDeleteCount() {
	m.Lock()
	m.DeleteCount++
	m.LastDelete = time.Now()
	m.Unlock()
}

// IncrementQueryCount increases the count of query operations and updates the timestamp of the last query operation.
func (m *Metrics) IncrementQueryCount() {
	m.Lock()
	m.QueryCount++
	m.LastQuery = time.Now()
	m.Unlock()
}

// IncrementCacheHits increases the count of successful cache retrievals.
func (m *Metrics) IncrementCacheHits() {
	m.Lock()
	m.CacheHits++
	m.Unlock()
}

// IncrementCacheMisses increases the count of unsuccessful cache retrievals.
func (m *Metrics) IncrementCacheMisses() {
	m.Lock()
	m.CacheMisses++
	m.Unlock()
}

// String returns a string representation of the Metrics structure in JSON format.
func (m *Metrics) String() string {
	m.RLock()
	defer m.RUnlock()
	metrics, _ := json.MarshalIndent(m, "", "  ")
	return string(metrics)
}
