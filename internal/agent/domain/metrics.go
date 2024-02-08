package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dnsoftware/go-metrics/internal/constants"
	"github.com/dnsoftware/go-metrics/internal/logger"
	"math/rand"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"sync"
	"syscall"
	"time"
)

//go:generate go run github.com/vektra/mockery/v2@v2.20.2 --name=AgentStorage
type AgentStorage interface {
	SetGauge(name string, value float64) error
	GetGauge(name string) (float64, error)

	SetCounter(name string, value int64) error
	GetCounter(name string) (int64, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.20.2 --name=MetricsSender
type MetricsSender interface {
	SendData(mType string, name string, value string) error
	SendDataBatch([]byte) error
}

type Flags interface {
	ReportInterval() int64
	PollInterval() int64
}

type Metrics struct {
	metrics        runtime.MemStats
	storage        AgentStorage
	sender         MetricsSender
	pollInterval   int64
	reportInterval int64
}

// MetricsItem структура для отправки json данных на сервер
type MetricsItem struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

var gaugeMetricsList []string = []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc", "RandomValue"}

func NewMetrics(storage AgentStorage, sender MetricsSender, flags Flags) Metrics {
	return Metrics{
		storage:        storage,
		sender:         sender,
		pollInterval:   flags.PollInterval(),
		reportInterval: flags.ReportInterval(),
	}
}

func (m *Metrics) Start() {
	var wg sync.WaitGroup

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// обновление метрик
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				fmt.Println("\nОбновление метрик завершено...")
				return
			default:
				m.updateMetrics()
				time.Sleep(time.Duration(m.pollInterval) * time.Second)
			}
		}

	}()

	// отправка метрик
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				fmt.Println("\nОтправка метрик завершена...")
				return
			default:
				time.Sleep(time.Duration(m.reportInterval) * time.Second)
				// в старом API было m.sendMetrics()
				m.sendMetricsBatch()
			}

		}
	}()

	wg.Wait()
	fmt.Println("Программа завершена!")
}

func (m *Metrics) updateMetrics() {

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	mCounter, err := m.storage.GetCounter(constants.PollCount)
	if err != nil {
		fmt.Println(err)
	}

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

			mCounter++

		} else if metricName == constants.RandomValue {
			m.storage.SetGauge(metricName, rng.Float64())
			mCounter++
		} else {
			// ... логируем ошибку, или еще что-то
			continue
		}
	}

	m.storage.SetCounter(constants.PollCount, mCounter)

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
	cValue, err := m.storage.GetCounter(constants.PollCount)
	if err != nil {
		cValue = 0
	}

	err = m.SendCounter(constants.PollCount, cValue)
	if err != nil {
		// обработка
		fmt.Println("Set PollCounter error: " + err.Error())
	}

	m.storage.SetCounter(constants.PollCount, 0)

}

// отправка метрик пакетом
func (m *Metrics) sendMetricsBatch() {

	var batch []MetricsItem

	// gauges
	for _, metricName := range gaugeMetricsList {
		val, err := m.storage.GetGauge(metricName)
		if err != nil {
			logger.Log().Error(err.Error())
			continue
		}

		batch = append(batch, MetricsItem{
			ID:    metricName,
			MType: constants.Gauge,
			Value: &val,
		})
	}

	// counters
	pollCount, err := m.storage.GetCounter(constants.PollCount)
	if err != nil {
		logger.Log().Error(err.Error())
		pollCount = 0
	}

	batch = append(batch, MetricsItem{
		ID:    constants.PollCount,
		MType: constants.Counter,
		Delta: &pollCount,
	})

	jsonData, err := json.Marshal(batch)

	err = m.sender.SendDataBatch(jsonData)
	if err != nil {
		logger.Log().Error(err.Error())
	}

}

func (m *Metrics) SendGauge(name string, value float64) error {

	err := m.sender.SendData(constants.Gauge, name, fmt.Sprintf("%f", value))
	if err != nil {
		return err
	}

	return nil
}

func (m *Metrics) SendCounter(name string, value int64) error {

	v := fmt.Sprintf("%d", value)
	err := m.sender.SendData(constants.Counter, name, v)
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
