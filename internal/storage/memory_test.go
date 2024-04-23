package storage

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetGetGauge(t *testing.T) {
	m := NewMemStorage()
	ctx := context.Background()

	m.SetGauge(ctx, "test", 123.456)
	val, err := m.GetGauge(ctx, "test")

	assert.NoError(t, err)
	assert.Equal(t, 123.456, val)
}

func TestSetGetCounter(t *testing.T) {
	m := NewMemStorage()
	ctx := context.Background()

	m.SetCounter(ctx, "test", 123)
	val, err := m.GetCounter(ctx, "test")

	assert.NoError(t, err)
	assert.Equal(t, int64(123), val)
}

func TestSetGetAll(t *testing.T) {
	m := NewMemStorage()
	ctx := context.Background()

	m.SetGauge(ctx, "Gauge", 123.456)
	m.SetCounter(ctx, "Counter", 123)
	gauges, counters, err := m.GetAll(ctx)
	fmt.Println(gauges, counters)

	assert.NoError(t, err)
	assert.Equal(t, 123.456, gauges["Gauge"])
	assert.Equal(t, int64(123), counters["Counter"])
}

func TestSetGetDump(t *testing.T) {
	m := NewMemStorage()
	ctx := context.Background()

	m.SetGauge(ctx, "Gauge", 123.456)
	m.SetCounter(ctx, "Counter", 123)
	dump, err := m.GetDump(ctx)
	fmt.Println(dump)

	assert.NoError(t, err)
	assert.Equal(t, `{"gauges":{"Gauge":123.456},"counters":{"Counter":123}}`, dump)
}

func TestRestoreFromDump(t *testing.T) {
	m := NewMemStorage()
	ctx := context.Background()

	err := m.RestoreFromDump(ctx, `{"gauges":{"Gauge":123.456},"counters":{"Counter":123}}`)
	assert.NoError(t, err)

	gauges, counters, err := m.GetAll(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 123.456, gauges["Gauge"])
	assert.Equal(t, int64(123), counters["Counter"])
}

func TestSetBatch(t *testing.T) {
	m := NewMemStorage()
	ctx := context.Background()

	batch := `[{"id":"Alloc","type":"gauge","value":343728},{"id":"BuckHashSys","type":"gauge","value":7321},{"id":"Frees","type":"gauge","value":268},{"id":"GCCPUFraction","type":"gauge","value":0},{"id":"GCSys","type":"gauge","value":1758064},{"id":"HeapAlloc","type":"gauge","value":343728},{"id":"HeapIdle","type":"gauge","value":2490368},{"id":"HeapInuse","type":"gauge","value":1179648},{"id":"HeapObjects","type":"gauge","value":1959},{"id":"HeapReleased","type":"gauge","value":2490368},{"id":"HeapSys","type":"gauge","value":3670016},{"id":"LastGC","type":"gauge","value":0},{"id":"Lookups","type":"gauge","value":0},{"id":"MCacheInuse","type":"gauge","value":4800},{"id":"MCacheSys","type":"gauge","value":15600},{"id":"MSpanInuse","type":"gauge","value":54400},{"id":"MSpanSys","type":"gauge","value":65280},{"id":"Mallocs","type":"gauge","value":2227},{"id":"NextGC","type":"gauge","value":4194304},{"id":"NumForcedGC","type":"gauge","value":0},{"id":"NumGC","type":"gauge","value":0},{"id":"OtherSys","type":"gauge","value":1126423},{"id":"PauseTotalNs","type":"gauge","value":0},{"id":"StackInuse","type":"gauge","value":524288},{"id":"StackSys","type":"gauge","value":524288},{"id":"Sys","type":"gauge","value":7166992},{"id":"TotalAlloc","type":"gauge","value":343728},{"id":"RandomValue","type":"gauge","value":0.5116380300334399},{"id":"TotalMemory","type":"gauge","value":33518669824},{"id":"FreeMemory","type":"gauge","value":3527917568},{"id":"CPUutilization1","type":"gauge","value":59.793814433156726},{"id":"CPUutilization2","type":"gauge","value":46.487603307343086},{"id":"CPUutilization3","type":"gauge","value":42.47422680367953},{"id":"CPUutilization4","type":"gauge","value":25.63559321940101},{"id":"PollCount","type":"counter","delta":62}]`

	err := m.SetBatch(ctx, []byte(batch))
	assert.NoError(t, err)

	assert.Equal(t, float64(343728), m.Gauges["Alloc"])
	assert.Equal(t, float64(2490368), m.Gauges["HeapIdle"])
	assert.Equal(t, float64(7166992), m.Gauges["Sys"])
	assert.Equal(t, float64(46.487603307343086), m.Gauges["CPUutilization2"])

	assert.Equal(t, int64(62), m.Counters["PollCount"])

	//	fmt.Println(m.Gauges, res)
}
