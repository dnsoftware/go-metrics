package domain

import (
	"context"
	"fmt"
	"github.com/dnsoftware/go-metrics/internal/agent/infrastructure"
	"github.com/dnsoftware/go-metrics/internal/constants"
	"github.com/dnsoftware/go-metrics/internal/storage"
	"github.com/shirou/gopsutil/v3/cpu"
	"strconv"
	"testing"
)

func BenchmarkSetBatch(b *testing.B) {
	m := storage.NewMemStorage()
	ctx := context.Background()
	batch := `[{"id":"Alloc","type":"gauge","value":343728},{"id":"BuckHashSys","type":"gauge","value":7321},{"id":"Frees","type":"gauge","value":268},{"id":"GCCPUFraction","type":"gauge","value":0},{"id":"GCSys","type":"gauge","value":1758064},{"id":"HeapAlloc","type":"gauge","value":343728},{"id":"HeapIdle","type":"gauge","value":2490368},{"id":"HeapInuse","type":"gauge","value":1179648},{"id":"HeapObjects","type":"gauge","value":1959},{"id":"HeapReleased","type":"gauge","value":2490368},{"id":"HeapSys","type":"gauge","value":3670016},{"id":"LastGC","type":"gauge","value":0},{"id":"Lookups","type":"gauge","value":0},{"id":"MCacheInuse","type":"gauge","value":4800},{"id":"MCacheSys","type":"gauge","value":15600},{"id":"MSpanInuse","type":"gauge","value":54400},{"id":"MSpanSys","type":"gauge","value":65280},{"id":"Mallocs","type":"gauge","value":2227},{"id":"NextGC","type":"gauge","value":4194304},{"id":"NumForcedGC","type":"gauge","value":0},{"id":"NumGC","type":"gauge","value":0},{"id":"OtherSys","type":"gauge","value":1126423},{"id":"PauseTotalNs","type":"gauge","value":0},{"id":"StackInuse","type":"gauge","value":524288},{"id":"StackSys","type":"gauge","value":524288},{"id":"Sys","type":"gauge","value":7166992},{"id":"TotalAlloc","type":"gauge","value":343728},{"id":"RandomValue","type":"gauge","value":0.5116380300334399},{"id":"TotalMemory","type":"gauge","value":33518669824},{"id":"FreeMemory","type":"gauge","value":3527917568},{"id":"CPUutilization1","type":"gauge","value":59.793814433156726},{"id":"CPUutilization2","type":"gauge","value":46.487603307343086},{"id":"CPUutilization3","type":"gauge","value":42.47422680367953},{"id":"CPUutilization4","type":"gauge","value":25.63559321940101},{"id":"PollCount","type":"counter","delta":62}]`

	for i := 0; i < b.N; i++ {
		err := m.SetBatch(ctx, []byte(batch))
		if err != nil {
			fmt.Errorf("Ошибка сохранения")
		}
	}
}

/*
	Суть нижеследующих бенчмарков и оптимизации кода в замене доступа к переменным через рефлексию на прямой доступ
	Уменьшение кол-ва аллокаций м увеличение быстродействия имеет место быть))

	UpdateMetricsReflect() - старый метод через рефлексию
	UpdateMetrics() - новый метод
*/

// go test -bench BenchmarkUpdateMetricsReflect -benchmem
// result:
// 38896             29369 ns/op            5810 B/op         55 allocs/op
func BenchmarkUpdateMetricsReflect(b *testing.B) {
	metrics := updateMetricsSetup()
	for i := 0; i < b.N; i++ {
		metrics.UpdateMetricsReflect()
	}
}

// go test -bench BenchmarkUpdateMetrics -benchmem
// result:
// 67407             18325 ns/op            5377 B/op          1 allocs/op
func BenchmarkUpdateMetrics(b *testing.B) {
	metrics := updateMetricsSetup()
	for i := 0; i < b.N; i++ {
		metrics.UpdateMetrics()
	}
}

type fl struct {
}

func updateMetricsSetup() Metrics {
	repository := storage.NewMemStorage()
	flg := fl{}
	sender := infrastructure.NewWebSender("http", &flg, constants.ApplicationJSON)

	gopcMetricsList := []string{constants.TotalMemory, constants.FreeMemory}

	cpuCount, _ := cpu.Counts(false)
	for i := 1; i <= cpuCount; i++ {
		gopcMetricsList = append(gopcMetricsList, constants.CPUutilization+strconv.Itoa(i))
	}

	metrics := NewMetrics(repository, &sender, &flg)

	return metrics
}

func (f *fl) RunAddr() string {
	return "localhost:8080"
}

func (f *fl) CryptoKey() string {
	return ""
}

func (f *fl) ReportInterval() int64 {
	return 10
}

func (f *fl) PollInterval() int64 {
	return 10
}

func (f *fl) RateLimit() int {
	return 10
}
