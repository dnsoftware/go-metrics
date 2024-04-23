// Package domain включает в себя бизнес логика агента. Получение метрик и отправка их на сервер
package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/dnsoftware/go-metrics/internal/constants"
	"github.com/dnsoftware/go-metrics/internal/logger"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// AgentStorage интерфейс хранилаща для агента.
type AgentStorage interface {
	// SetGauge сохранение метрики типа gauge в хранилище.
	// Параметры: name - название метрики, value - ее значение.
	SetGauge(ctx context.Context, name string, value float64) error

	// GetGauge получение значения метрики типа gauge из хранилища.
	// Параметры: name - название метрики.
	GetGauge(ctx context.Context, name string) (float64, error)

	// SetCounter сохранение метрики типа counter в хранилище.
	// Параметры: name - название метрики, value - ее значение.
	SetCounter(ctx context.Context, name string, value int64) error

	// GetCounter получение значения метрики типа counter из хранилища.
	// Параметры: name - название метрики.
	GetCounter(ctx context.Context, name string) (int64, error)
}

// MetricsSender отправка метрик на сервер.
type MetricsSender interface {
	// SendData отправка одной метрики на сервер.
	SendData(ctx context.Context, mType string, name string, value string) error

	// SendDataBatch отправка метрик на сервер пакетом. Параметр jsonData - данные по
	// метрикам в формате json в виде среза байт
	SendDataBatch(ctx context.Context, jsonData []byte) error
}

// Flags получение значения флагов командной строки запуска агента
type Flags interface {
	ReportInterval() int64
	PollInterval() int64
	RateLimit() int
}

// Metrics основная структура агента. Получение, промежуточное сохранение, отправка данных на сервер.
type Metrics struct {
	metrics         runtime.MemStats
	storage         AgentStorage
	sender          MetricsSender
	pollInterval    int64
	reportInterval  int64
	gopcMetricsList []string
	rateLimit       int
}

// MetricsItem структура для отправки json данных на сервер
type MetricsItem struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

// gaugeMetricsList названия всех доступных gauge метрик
var gaugeMetricsList = []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc", "RandomValue"}

func NewMetrics(storage AgentStorage, sender MetricsSender, flags Flags) Metrics {

	gopcMetricsList := []string{constants.TotalMemory, constants.FreeMemory}

	cpuCount, _ := cpu.Counts(false)
	for i := 1; i <= cpuCount; i++ {
		gopcMetricsList = append(gopcMetricsList, constants.CPUutilization+strconv.Itoa(i))
	}

	return Metrics{
		storage:         storage,
		sender:          sender,
		pollInterval:    flags.PollInterval(),
		reportInterval:  flags.ReportInterval(),
		rateLimit:       flags.RateLimit(),
		gopcMetricsList: gopcMetricsList,
	}
}

// Start старт работы агента - запуск горутин по получению и отправке метрик.
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
				m.UpdateMetrics()
				time.Sleep(time.Duration(m.pollInterval) * time.Second)
			}
		}
	}()

	// для gopcutils
	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				fmt.Println("\nОбновление метрик gopcutils завершено...")
				return
			default:
				m.UpdateGopcMetrics()
				time.Sleep(time.Duration(m.pollInterval) * time.Second)
			}
		}
	}()

	// отправка метрик

	// создаем буферизованный канал для принятия задач в воркер
	jobsCh := make(chan []byte, constants.ChannelCap)

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				fmt.Println("\nПодготовка пакетов на отправку завершена...")
				return
			default:
				time.Sleep(time.Duration(m.reportInterval) * time.Second)
				// в старом API было m.sendMetrics()
				m.sendMetricsBatch(ctx, jobsCh)
			}
		}
	}()

	// отправка пакетов, поступающих в очередь
	rateLimitChan := make(chan struct{}, constants.RateLimit)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(rateLimitChan)

		for {
			select {
			case <-ctx.Done():
				fmt.Println("\nОтправка метрик завершена...") ///
				return
			case job := <-jobsCh:
				rateLimitChan <- struct{}{}
				go m.worker(job, rateLimitChan)
			}
		}
	}()

	wg.Wait()

	fmt.Println("\nПрограмма завершена!")
}

// UpdateMetricsReflect сохранение метрик в базу с использованием рефлексии.
func (m *Metrics) UpdateMetricsReflect() {
	ctx := context.Background()

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	mCounter, err := m.storage.GetCounter(ctx, constants.PollCount)
	if err != nil {
		logger.Log().Error(err.Error())
	}

	runtime.ReadMemStats(&m.metrics)

	for _, metricName := range gaugeMetricsList {
		mType, ok := reflect.TypeOf(&m.metrics).Elem().FieldByName(metricName)
		if ok { // если поле найдено
			tp := mType.Type.Kind().String()
			metricValue := reflect.ValueOf(&m.metrics).Elem().FieldByName(metricName)

			switch tp {
			case "uint64":
				m.storage.SetGauge(ctx, metricName, float64(metricValue.Uint()))
			case "uint32":
				m.storage.SetGauge(ctx, metricName, float64(metricValue.Uint()))
			case "float64":
				m.storage.SetGauge(ctx, metricName, metricValue.Float())
			default:
				// действия при неучтенном типе
				logger.Log().Error("неучтенный тип метрики: " + tp)
			}

			mCounter++
		} else if metricName == constants.RandomValue {
			m.storage.SetGauge(ctx, metricName, rnd.Float64())

			mCounter++
		} else {
			logger.Log().Warn("Metric not found: " + metricName)

			continue
		}
	}

	m.storage.SetCounter(ctx, constants.PollCount, mCounter)
}

// UpdateMetrics сохранение метрик в базу явным образом.
// работает быстрее
func (m *Metrics) UpdateMetrics() {
	ctx := context.Background()

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	mCounter, err := m.storage.GetCounter(ctx, constants.PollCount)
	if err != nil {
		logger.Log().Error(err.Error())
	}

	runtime.ReadMemStats(&m.metrics)

	m.storage.SetGauge(ctx, "Alloc", float64(m.metrics.Alloc))
	m.storage.SetGauge(ctx, "TotalAlloc", float64(m.metrics.TotalAlloc))
	m.storage.SetGauge(ctx, "Sys", float64(m.metrics.Sys))
	m.storage.SetGauge(ctx, "Lookups", float64(m.metrics.Lookups))
	m.storage.SetGauge(ctx, "Mallocs", float64(m.metrics.Mallocs))
	m.storage.SetGauge(ctx, "Frees", float64(m.metrics.Frees))
	m.storage.SetGauge(ctx, "HeapAlloc", float64(m.metrics.HeapAlloc))
	m.storage.SetGauge(ctx, "HeapSys", float64(m.metrics.HeapSys))
	m.storage.SetGauge(ctx, "HeapIdle", float64(m.metrics.HeapIdle))
	m.storage.SetGauge(ctx, "HeapInuse", float64(m.metrics.HeapInuse))
	m.storage.SetGauge(ctx, "HeapReleased", float64(m.metrics.HeapReleased))
	m.storage.SetGauge(ctx, "HeapObjects", float64(m.metrics.HeapObjects))
	m.storage.SetGauge(ctx, "StackInuse", float64(m.metrics.StackInuse))
	m.storage.SetGauge(ctx, "StackSys", float64(m.metrics.StackSys))
	m.storage.SetGauge(ctx, "MSpanInuse", float64(m.metrics.MSpanInuse))
	m.storage.SetGauge(ctx, "MSpanSys", float64(m.metrics.MSpanSys))
	m.storage.SetGauge(ctx, "MCacheInuse", float64(m.metrics.MCacheInuse))
	m.storage.SetGauge(ctx, "MCacheSys", float64(m.metrics.MCacheSys))
	m.storage.SetGauge(ctx, "BuckHashSys", float64(m.metrics.BuckHashSys))
	m.storage.SetGauge(ctx, "GCSys", float64(m.metrics.GCSys))
	m.storage.SetGauge(ctx, "OtherSys", float64(m.metrics.OtherSys))
	m.storage.SetGauge(ctx, "NextGC", float64(m.metrics.NextGC))
	m.storage.SetGauge(ctx, "LastGC", float64(m.metrics.LastGC))
	m.storage.SetGauge(ctx, "PauseTotalNs", float64(m.metrics.PauseTotalNs))
	m.storage.SetGauge(ctx, "NumGC", float64(m.metrics.NumGC))
	m.storage.SetGauge(ctx, "NumForcedGC", float64(m.metrics.NumForcedGC))
	m.storage.SetGauge(ctx, "GCCPUFraction", m.metrics.GCCPUFraction)
	m.storage.SetGauge(ctx, constants.RandomValue, rnd.Float64())
	mCounter += 28

	m.storage.SetCounter(ctx, constants.PollCount, mCounter)
}

// updateGopcMetrics получение системных и аппаратных метрик и сохранение их в базу.
func (m *Metrics) UpdateGopcMetrics() {
	ctx := context.Background()

	mCounter, err := m.storage.GetCounter(ctx, constants.PollCount)
	if err != nil {
		logger.Log().Error(err.Error())
	}

	vm, _ := mem.VirtualMemory()

	m.storage.SetGauge(ctx, constants.TotalMemory, float64(vm.Total))
	m.storage.SetGauge(ctx, constants.FreeMemory, float64(vm.Free))
	mCounter += 2

	cc, _ := cpu.Percent(time.Second*time.Duration(constants.CPUIntervalUtilization), true)
	for key, val := range cc {
		m.storage.SetGauge(ctx, constants.CPUutilization+strconv.Itoa(key+1), val)
		mCounter++
	}

	m.storage.SetCounter(ctx, constants.PollCount, mCounter)
}

// sendMetrics отправка метрик на сервер
func (m *Metrics) sendMetrics() {
	ctx, cancel := context.WithTimeout(context.Background(), constants.HTTPContextTimeout)
	defer cancel()

	// gauge
	allGaugeMetrics := append(gaugeMetricsList, m.gopcMetricsList...)
	for _, metricName := range allGaugeMetrics {
		val, err := m.storage.GetGauge(ctx, metricName)
		if err != nil {
			logger.Log().Error(err.Error())
			continue
		}

		err = m.SendGauge(ctx, metricName, val)
		if err != nil {
			logger.Log().Error(err.Error())
			continue
		}
	}

	// counter
	cValue, err := m.storage.GetCounter(ctx, constants.PollCount)
	if err != nil {
		cValue = 0
	}

	err = m.SendCounter(ctx, constants.PollCount, cValue)
	if err != nil {
		logger.Log().Error("Set PollCounter error: " + err.Error())
	}

	_ = m.storage.SetCounter(ctx, constants.PollCount, 0)
}

// worker воркер по пакетной отправке метрик на сервер.
func (m *Metrics) worker(job []byte, rateLimitChan chan struct{}) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.HTTPContextTimeout)
	defer cancel()

	err := m.sender.SendDataBatch(ctx, job)
	if err != nil {
		logger.Log().Error("Send job error: " + err.Error())
	}

	<-rateLimitChan
}

// отправка метрик мини-пакетами
func (m *Metrics) sendMetricsBatch(ctx context.Context, jobsCh chan []byte) {
	ctx, cancel := context.WithTimeout(ctx, constants.HTTPContextTimeout)
	defer cancel()

	select {
	case <-ctx.Done():
		fmt.Println("\nОтправка пакетов завершена...") ///
		return
	default:
	}

	var batch []MetricsItem

	// gauges
	allGaugeMetrics := append(gaugeMetricsList, m.gopcMetricsList...)
	for _, metricName := range allGaugeMetrics {
		val, err := m.storage.GetGauge(ctx, metricName)
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
	pollCount, err := m.storage.GetCounter(ctx, constants.PollCount)
	if err != nil {
		logger.Log().Error(err.Error())

		pollCount = 0
	}

	batch = append(batch, MetricsItem{
		ID:    constants.PollCount,
		MType: constants.Counter,
		Delta: &pollCount,
	})

	// отправляем задачи, упакованные в мелкие пакеты, воркерам
	var (
		miniBatch []MetricsItem
		batchData []byte
	)

	i := 0

	for _, mi := range batch {
		miniBatch = append(miniBatch, mi)
		i++

		if i == constants.BatchItemCount {
			batchData, err = json.Marshal(miniBatch)
			if err != nil {
				logger.Log().Error(err.Error())
				return
			}

			jobsCh <- batchData

			i = 0
			miniBatch = nil
		}
	}

	if len(miniBatch) > 0 {
		batchData, err = json.Marshal(miniBatch)
		if err != nil {
			logger.Log().Error(err.Error())
			return
		}
		jobsCh <- batchData
	}

}

// SendGauge отправляет одну метрику типа gauge на сервер
func (m *Metrics) SendGauge(ctx context.Context, name string, value float64) error {
	err := m.sender.SendData(ctx, constants.Gauge, name, fmt.Sprintf("%f", value))
	if err != nil {
		return err
	}

	return nil
}

// SendCounter отправляет одну метрику типа counter на сервер
func (m *Metrics) SendCounter(ctx context.Context, name string, value int64) error {
	v := strconv.FormatInt(value, 10)

	err := m.sender.SendData(ctx, constants.Counter, name, v)
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
