package collector

type ServerStorage interface {
	SetGauge(name string, value float64)
	GetGauge(name string) (float64, error)

	SetCounter(name string, value int64)
	GetCounter(name string) (int64, error)
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
	return c.storage.GetCounter(metricName)
}
