package collector

type Collector struct {
}

type MemStorage struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

func NewCollector() *Collector {

	collector := &Collector{}

	return collector
}

// Допустимые gauge метрики
func possibleGauges() []string {
	return []string{""}
}

// Допустимые counter метрики
func possibleCounters() []string {
	return []string{""}
}

func (c *Collector) UpdateGauge(metricName string, metricValue float64) {}

func (c *Collector) UpdateCounter(metricName string, metricValue int64) {}
