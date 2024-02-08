package storage

import (
	"encoding/json"
	"errors"
	"github.com/dnsoftware/go-metrics/internal/constants"
	"sync"
)

type MemStorage struct {
	storageType string
	mutex       sync.Mutex
	Gauges      map[string]float64 `json:"gauges"`
	Counters    map[string]int64   `json:"counters"`
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		storageType: constants.Memory,
		Gauges:      make(map[string]float64),
		Counters:    make(map[string]int64),
	}
}

func (m *MemStorage) SetGauge(name string, value float64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.Gauges[name] = value

	return nil
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

func (m *MemStorage) SetCounter(name string, value int64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.Counters[name] = value

	return nil
}

func (m *MemStorage) SetBatch(batch []byte) error {
	var metrics []Metrics

	m.mutex.Lock()
	defer m.mutex.Unlock()

	err := json.Unmarshal(batch, &metrics)
	if err != nil {
		return err
	}

	for _, mt := range metrics {
		if mt.MType == constants.Gauge {
			m.Gauges[mt.ID] = *mt.Value
		}

		if mt.MType == constants.Counter {
			m.Counters[mt.ID] = *mt.Delta
		}
	}

	return nil
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
func (m *MemStorage) GetAll() (map[string]float64, map[string]int64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.Gauges, m.Counters, nil
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

func (m *MemStorage) Type() string {
	return m.storageType
}

func (m *MemStorage) Health() bool {
	return true
}
