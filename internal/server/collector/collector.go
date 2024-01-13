package collector

import (
	"errors"
	"fmt"
)

type ServerStorage interface {
	SetGauge(name string, value float64)
	GetGauge(name string) (float64, error)

	SetCounter(name string, value int64)
	GetCounter(name string) (int64, error)

	GetAll() (map[string]float64, map[string]int64)
}

type Collector struct {
	storage ServerStorage
}

const (
	gaugeSelector   string = "gauge"
	counterSelector string = "counter"
)

var gaugeMetricsList []string = []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc", "RandomValue"}

func NewCollector(storage ServerStorage) *Collector {

	collector := &Collector{
		storage: storage,
	}

	return collector
}

// проверка на допустимую метрику
func (c *Collector) isMetric(mType string, name string) bool {

	if mType == gaugeSelector {
		for _, val := range gaugeMetricsList {
			if val == name {
				return true
			}
		}
	}

	if mType == counterSelector {
		if name == "PollCount" {
			return true
		}
	}

	return false
}

func (c *Collector) SetGaugeMetric(metricName string, metricValue float64) error {

	/** // проверка на допустимость метрики убрана
	if !c.isMetric(gaugeSelector, metricName) {
		return errors.New("invalid metric")
	}
	/**/

	c.storage.SetGauge(metricName, metricValue)

	return nil
}
func (c *Collector) GetGaugeMetric(metricName string) (float64, error) {

	//if !c.isMetric(gaugeSelector, metricName) {
	//	return 0, errors.New("invalid metric")
	//}

	return c.storage.GetGauge(metricName)
}

func (c *Collector) SetCounterMetric(metricName string, metricValue int64) error {

	/** // проверка на допустимость метрики убрана
	if !c.isMetric(counterSelector, metricName) {
		return errors.New("invalid metric")
	}
	/**/

	c.storage.SetCounter(metricName, metricValue)

	return nil
}
func (c *Collector) GetCounterMetric(metricName string) (int64, error) {

	//if !c.isMetric(counterSelector, metricName) {
	//	return 0, errors.New("invalid metric")
	//}

	return c.storage.GetCounter(metricName)
}

// получение метрики в текстовом виде
func (c *Collector) GetMetric(metricType string, metricName string) (string, error) {

	var valStr string

	switch metricType {
	case "gauge":
		val, err := c.GetGaugeMetric(metricName)
		if err != nil {
			return "", err
		}
		valStr = fmt.Sprintf("%f", val)

	case "counter":
		val, err := c.GetCounterMetric(metricName)
		if err != nil {
			return "", err
		}
		valStr = fmt.Sprintf("%v", val)
	default:
		return "", errors.New("bad metric type")
	}

	return valStr, nil

}

// все метрики списком
func (c *Collector) GetAll() (string, error) {
	gauges, counters := c.storage.GetAll()

	mList := ""
	for key, val := range gauges {
		mList = mList + key + ": " + fmt.Sprintf("%f", val) + "\n"
	}

	for key, val := range counters {
		mList = mList + key + ": " + fmt.Sprintf("%v", val) + "\n"
	}

	return mList, nil
}
