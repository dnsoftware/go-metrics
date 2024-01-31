package storage

import (
	"encoding/json"
	"errors"
	"sync"
)

type MemStorage struct {
	mutex    sync.Mutex
	Gauges   map[string]float64 `json:"gauges"`
	Counters map[string]int64   `json:"counters"`
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
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

// получение json дампа
func (m *MemStorage) GetDump() (string, error) {

	m.mutex.Lock()
	defer m.mutex.Unlock()

	data, err := json.Marshal(m)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// восстановление из json дампа
func (m *MemStorage) RestoreFromDump(dump string) error {

	m.mutex.Lock()
	defer m.mutex.Unlock()

	err := json.Unmarshal([]byte(dump), m)
	if err != nil {
		return err
	}

	return nil
}
