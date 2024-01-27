package storage

import (
	"errors"
	"sync"
)

type MemStorage struct {
	mutex    sync.Mutex
	Gauges   map[string]float64
	Counters map[string]int64
}

func NewMemStorage() MemStorage {
	return MemStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}
}

func (m *MemStorage) SetGauge(name string, value float64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.Gauges[name] = value
}

func (m *MemStorage) GetGauge(name string) (float64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if value, ok := m.Gauges[name]; ok {
		return value, nil
	} else {
		return 0, errors.New("no such metric")
	}

}

func (m *MemStorage) SetCounter(name string, value int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.Counters[name] = value

}

func (m *MemStorage) GetCounter(name string) (int64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if value, ok := m.Counters[name]; ok {
		return value, nil
	} else {
		return 0, errors.New("no such metric")
	}
}

// возврат карт gauge и counters
func (m *MemStorage) GetAll() (map[string]float64, map[string]int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.Gauges, m.Counters
}
