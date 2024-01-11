package storage

type Storage interface {
	SetGauge(name string, value float64)
	GetGauge(name string) (float64, error)

	SetCounter(name string, value int64)
	GetCounter(name string) (int64, error)
}
