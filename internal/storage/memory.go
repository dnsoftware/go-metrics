package storage

import "errors"

type MemStorage struct {
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

	m.Gauges[name] = value

	return
}

func (m *MemStorage) GetGauge(name string) (float64, error) {

	if value, ok := m.Gauges[name]; ok {
		return value, nil
	} else {
		return 0, errors.New("no such metric")
	}

}

func (m *MemStorage) SetCounter(name string, value int64) {
	m.Counters[name] = value

	return
}

func (m *MemStorage) GetCounter(name string) (int64, error) {
	if value, ok := m.Counters[name]; ok {
		return value, nil
	} else {
		return 0, errors.New("no such metric")
	}
}
