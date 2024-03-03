package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dnsoftware/go-metrics/internal/constants"
	"github.com/dnsoftware/go-metrics/internal/logger"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"math/rand"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"
)

//go:generate go run github.com/vektra/mockery/v2@v2.20.2 --name=AgentStorage
type AgentStorage interface {
	SetGauge(ctx context.Context, name string, value float64) error
	GetGauge(ctx context.Context, name string) (float64, error)

	SetCounter(ctx context.Context, name string, value int64) error
	GetCounter(ctx context.Context, name string) (int64, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.20.2 --name=MetricsSender
type MetricsSender interface {
	SendData(ctx context.Context, mType string, name string, value string) error
	SendDataBatch(ctx context.Context, jsonData []byte) error
}

type Flags interface {
	ReportInterval() int64
	PollInterval() int64
	RateLimit() int
}

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
				m.updateGopcMetrics()
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

func (m *Metrics) updateMetrics() {
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

func (m *Metrics) updateGopcMetrics() {
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

func (m *Metrics) SendGauge(ctx context.Context, name string, value float64) error {
	err := m.sender.SendData(ctx, constants.Gauge, name, fmt.Sprintf("%f", value))
	if err != nil {
		return err
	}

	return nil
}

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
