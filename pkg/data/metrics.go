package data

import (
	"encoding/json"
	"sync"
	"time"
)

type Metrics struct {
	sync.RWMutex
	InsertCount int
	UpdateCount int
	DeleteCount int
	QueryCount  int
	CacheHits   int
	CacheMisses int
	LastInsert  time.Time
	LastUpdate  time.Time
	LastDelete  time.Time
	LastQuery   time.Time
}

func NewMetrics() *Metrics {
	return &Metrics{}
}

func (m *Metrics) IncrementInsertCount() {
	m.Lock()
	m.InsertCount++
	m.LastInsert = time.Now()
	m.Unlock()
}

func (m *Metrics) IncrementUpdateCount() {
	m.Lock()
	m.UpdateCount++
	m.LastUpdate = time.Now()
	m.Unlock()
}

func (m *Metrics) IncrementDeleteCount() {
	m.Lock()
	m.DeleteCount++
	m.LastDelete = time.Now()
	m.Unlock()
}

func (m *Metrics) IncrementQueryCount() {
	m.Lock()
	m.QueryCount++
	m.LastQuery = time.Now()
	m.Unlock()
}

func (m *Metrics) IncrementCacheHits() {
	m.Lock()
	m.CacheHits++
	m.Unlock()
}

func (m *Metrics) IncrementCacheMisses() {
	m.Lock()
	m.CacheMisses++
	m.Unlock()
}

func (m *Metrics) String() string {
	m.RLock()
	defer m.RUnlock()
	metrics, _ := json.MarshalIndent(m, "", "  ")
	return string(metrics)
}
