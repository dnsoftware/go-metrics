package domain

import (
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"time"
)

//go:generate go run github.com/vektra/mockery/v2@v2.20.2 --name=AgentStorage
type AgentStorage interface {
	SetGauge(name string, value float64)
	GetGauge(name string) (float64, error)

	SetCounter(name string, value int64)
	GetCounter(name string) (int64, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.20.2 --name=MetricsSender
type MetricsSender interface {
	SendData(mType string, name string, value string) error
}

type Metrics struct {
	metrics runtime.MemStats
	storage AgentStorage
	sender  MetricsSender
}

const (
	pollInterval   int64 = 2  // интервал обновления метрик
	reportInterval int64 = 10 // интервал отправки метрик на сервер
)

var gaugeMetricsList []string = []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc", "RandomValue"}

func NewMetrics(storage AgentStorage, sender MetricsSender) Metrics {
	return Metrics{
		storage: storage,
		sender:  sender,
	}
}

func (m *Metrics) Start() {

	// обновление метрик
	go func() {
		for {
			m.updateMetrics()
			time.Sleep(time.Duration(pollInterval) * time.Second)
		}
	}()

	// отправка метрик
	go func() {
		for {
			time.Sleep(time.Duration(reportInterval) * time.Second)
			m.sendMetrics()
		}
	}()

}

func (m *Metrics) updateMetrics() {

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	runtime.ReadMemStats(&m.metrics)
	for _, metricName := range gaugeMetricsList {
		mType, ok := reflect.TypeOf(&m.metrics).Elem().FieldByName(metricName)
		if ok { // если поле найдено
			tp := mType.Type.Kind().String()
			metricValue := reflect.ValueOf(&m.metrics).Elem().FieldByName(metricName)

			switch tp {
			case "uint64":
				m.storage.SetGauge(metricName, float64(metricValue.Uint()))
			case "uint32":
				m.storage.SetGauge(metricName, float64(metricValue.Uint()))
			case "float64":
				m.storage.SetGauge(metricName, metricValue.Float())
			default:
				// действия при неучтенном типе
			}

		} else if metricName == "RandomValue" {
			m.storage.SetGauge(metricName, rng.Float64())
		} else {
			// ... логируем ошибку, или еще что-то
			continue
		}
	}

}

func (m *Metrics) sendMetrics() {

	// gauge
	for _, metricName := range gaugeMetricsList {
		val, err := m.storage.GetGauge(metricName)
		if err != nil {
			// обработка ошибки
			continue
		}

		err = m.SendGauge(metricName, val)
		if err != nil {
			// обработка ошибки
			continue
		}
	}

	// counter
	cValue, err := m.storage.GetCounter("PollCount")
	if err != nil {
		cValue = 0
	}
	cValue++
	m.storage.SetCounter("PollCount", cValue)

	err = m.SendCounter("PollCount", cValue)
	if err != nil {
		// обработка
		fmt.Println("Set PollCounter error: " + err.Error())
	}

}

func (m *Metrics) SendGauge(name string, value float64) error {

	err := m.sender.SendData("gauge", name, fmt.Sprintf("%f", value))
	if err != nil {
		return err
	}

	return nil
}

func (m *Metrics) SendCounter(name string, value int64) error {

	v := fmt.Sprintf("%d", value)
	err := m.sender.SendData("counter", name, v)
	if err != nil {
		return err
	}

	return nil
}

// проверка на допустимую метрику gauge
func (m *Metrics) IsGauge(name string) bool {
	for _, val := range gaugeMetricsList {
		if val == name {
			return true
		}
	}

	return false
}
