package collector

import (
	"context"
	"errors"
	"fmt"
	"github.com/dnsoftware/go-metrics/internal/constants"
	"github.com/dnsoftware/go-metrics/internal/logger"
	"github.com/dnsoftware/go-metrics/internal/server/config"
	"strconv"
	"time"
)

type ServerStorage interface {
	SetGauge(ctx context.Context, name string, value float64) error
	SetCounter(ctx context.Context, name string, value int64) error
	SetBatch(ctx context.Context, batch []byte) error

	GetGauge(ctx context.Context, name string) (float64, error)
	GetCounter(ctx context.Context, name string) (int64, error)
	GetAll(ctx context.Context) (map[string]float64, map[string]int64, error)

	GetDump(ctx context.Context) (string, error)
	RestoreFromDump(ctx context.Context, dump string) error

	DatabasePing(ctx context.Context) bool
}

type BackupStorage interface {
	Save(dump string) error
	Load() (string, error)
}

type Collector struct {
	cfg           *config.ServerConfig
	storage       ServerStorage
	backupStorage BackupStorage
}

var gaugeMetricsList = []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc", "RandomValue"}

func NewCollector(cfg *config.ServerConfig, storage ServerStorage, backupStorage BackupStorage) (*Collector, error) {
	collector := &Collector{
		cfg:           cfg,
		storage:       storage,
		backupStorage: backupStorage,
	}

	// Загружаем сохраненную базу, если нужно
	if cfg.RestoreSaved {
		err := collector.loadFromDump()
		if err != nil {
			return nil, err
		}
	}

	collector.startBackup()

	return collector, nil
}

// проверка на допустимую метрику
func (c *Collector) isMetric(mType string, name string) bool {

	if mType == constants.Gauge {
		for _, val := range gaugeMetricsList {
			if val == name {
				return true
			}
		}
	}

	if mType == constants.Counter {
		if name == constants.PollCount {
			return true
		}
	}

	return false
}

func (c *Collector) SetGaugeMetric(ctx context.Context, metricName string, metricValue float64) error {
	err := c.storage.SetGauge(ctx, metricName, metricValue)
	if err != nil {
		return err
	}

	// если бэкап синхронный и указан файл
	if c.cfg.StoreInterval == constants.BackupPeriodSync && c.cfg.FileStoragePath != "" {
		err = c.generateDump()
		if err != nil {
			return err
		}
	}

	return nil
}
func (c *Collector) GetGaugeMetric(ctx context.Context, metricName string) (float64, error) {
	return c.storage.GetGauge(ctx, metricName)
}

// Прибавляем к уже существующему значению
func (c *Collector) SetCounterMetric(ctx context.Context, metricName string, metricValue int64) error {
	oldVal, _ := c.storage.GetCounter(ctx, metricName)
	newVal := oldVal + metricValue

	err := c.storage.SetCounter(ctx, metricName, newVal)
	if err != nil {
		return err
	}

	// если бэкап синхронный и указан файл
	if c.cfg.StoreInterval == constants.BackupPeriodSync && c.cfg.FileStoragePath != "" {
		errB := c.generateDump()
		if errB != nil {
			return errB
		}
	}

	return nil
}

func (c *Collector) SetBatchMetrics(ctx context.Context, batch []byte) error {
	return c.storage.SetBatch(ctx, batch)
}

func (c *Collector) GetCounterMetric(ctx context.Context, metricName string) (int64, error) {
	return c.storage.GetCounter(ctx, metricName)
}

// получение метрики в текстовом виде
func (c *Collector) GetMetric(ctx context.Context, metricType string, metricName string) (string, error) {
	var valStr string

	switch metricType {
	case constants.Gauge:
		val, err := c.GetGaugeMetric(ctx, metricName)
		if err != nil {
			return "", err
		}

		valStr = strconv.FormatFloat(val, 'f', -1, 64)

	case constants.Counter:
		val, err := c.GetCounterMetric(ctx, metricName)
		if err != nil {
			return "", err
		}

		valStr = strconv.FormatInt(val, 10)
	default:
		return "", errors.New("bad metric type")
	}

	return valStr, nil
}

// GetAll все метрики списком
func (c *Collector) GetAll(ctx context.Context) (string, error) {
	gauges, counters, err := c.storage.GetAll(ctx)
	if err != nil {
		return "", err
	}

	mList := ""
	for key, val := range gauges {
		mList = mList + key + ": " + fmt.Sprintf("%f", val) + "\n"
	}

	for key, val := range counters {
		mList = mList + key + ": " + strconv.FormatInt(val, 10) + "\n"
	}

	return mList, nil
}

// сохранение дампа в файл
func (c *Collector) generateDump() error {
	ctx, cancel := context.WithTimeout(context.Background(), constants.DBContextTimeout)
	defer cancel()

	dump, err := c.storage.GetDump(ctx)
	if err != nil {
		logger.Log().Error(err.Error())
		return err
	}

	err = c.backupStorage.Save(dump)
	if err != nil {
		logger.Log().Error(err.Error())
		return err
	}

	return nil
}

// загрузка данных из дампа
func (c *Collector) loadFromDump() error {
	ctx, cancel := context.WithTimeout(context.Background(), constants.DBContextTimeout)
	defer cancel()

	dump, err := c.backupStorage.Load()
	if err != nil {
		logger.Log().Error(err.Error())
		return err
	}

	// пустой файл
	if len(dump) == 0 {
		return nil
	}

	err = c.storage.RestoreFromDump(ctx, dump)
	if err != nil {
		logger.Log().Error(err.Error())
		return err
	}

	return nil
}

// периодическое сохранение метрик
func (c *Collector) startBackup() {
	// если обновление синхронное - не запускаем периодическое обновление
	if c.cfg.StoreInterval == constants.BackupPeriodSync {
		return
	}

	// если файл не указан - не запускаем сохранение на диск
	if c.cfg.FileStoragePath == "" {
		return
	}

	backupPeriod := time.Duration(c.cfg.StoreInterval) * time.Second

	go func() {
		for {
			time.Sleep(backupPeriod)

			err := c.generateDump()
			if err != nil {
				logger.Log().Error(err.Error())
			}
		}
	}()
}

// DatabasePing проверка работоспособности СУБД
func (c *Collector) DatabasePing(ctx context.Context) bool {
	return c.storage.DatabasePing(ctx)
}
