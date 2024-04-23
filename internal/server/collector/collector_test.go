package collector

import (
	"context"
	"testing"

	"github.com/dnsoftware/go-metrics/internal/constants"

	mock_collector "github.com/dnsoftware/go-metrics/internal/server/collector/mocks"
	"github.com/dnsoftware/go-metrics/internal/server/config"
	"github.com/dnsoftware/go-metrics/internal/storage"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewCollector(t *testing.T) {
	_, err := setup(t)
	assert.NoError(t, err)
}

func TestSetGetGaugeMetric(t *testing.T) {
	ctx := context.Background()

	c, err := setup(t)
	assert.NoError(t, err)

	testVal := 123.456
	err = c.SetGaugeMetric(ctx, "Alloc", testVal)
	assert.NoError(t, err)

	m, err := c.GetGaugeMetric(ctx, "Alloc")
	assert.Equal(t, testVal, m)
}

func TestSetGetCounterMetric(t *testing.T) {
	ctx := context.Background()

	c, err := setup(t)
	assert.NoError(t, err)

	testVal := int64(123)
	err = c.SetCounterMetric(ctx, constants.PollCount, testVal)
	assert.NoError(t, err)

	m, err := c.GetCounterMetric(ctx, constants.PollCount)
	assert.Equal(t, testVal, m)
}

func TestCollector_IsMetric(t *testing.T) {
	c, _ := setup(t)

	isMetric := c.IsMetric("gauge", "Alloc")
	assert.Equal(t, true, isMetric)

	isMetric = c.IsMetric("counter", constants.PollCount)
	assert.Equal(t, true, isMetric)

	isMetric = c.IsMetric("counterBad", constants.PollCount)
	assert.Equal(t, false, isMetric)

}

func TestCollector_SetBatchMetrics(t *testing.T) {
	ctx := context.Background()
	batch := `[{"id":"Alloc","type":"gauge","value":343728},{"id":"BuckHashSys","type":"gauge","value":7321},{"id":"Frees","type":"gauge","value":268},{"id":"GCCPUFraction","type":"gauge","value":0},{"id":"GCSys","type":"gauge","value":1758064},{"id":"HeapAlloc","type":"gauge","value":343728},{"id":"HeapIdle","type":"gauge","value":2490368},{"id":"HeapInuse","type":"gauge","value":1179648},{"id":"HeapObjects","type":"gauge","value":1959},{"id":"HeapReleased","type":"gauge","value":2490368},{"id":"HeapSys","type":"gauge","value":3670016},{"id":"LastGC","type":"gauge","value":0},{"id":"Lookups","type":"gauge","value":0},{"id":"MCacheInuse","type":"gauge","value":4800},{"id":"MCacheSys","type":"gauge","value":15600},{"id":"MSpanInuse","type":"gauge","value":54400},{"id":"MSpanSys","type":"gauge","value":65280},{"id":"Mallocs","type":"gauge","value":2227},{"id":"NextGC","type":"gauge","value":4194304},{"id":"NumForcedGC","type":"gauge","value":0},{"id":"NumGC","type":"gauge","value":0},{"id":"OtherSys","type":"gauge","value":1126423},{"id":"PauseTotalNs","type":"gauge","value":0},{"id":"StackInuse","type":"gauge","value":524288},{"id":"StackSys","type":"gauge","value":524288},{"id":"Sys","type":"gauge","value":7166992},{"id":"TotalAlloc","type":"gauge","value":343728},{"id":"RandomValue","type":"gauge","value":0.5116380300334399},{"id":"TotalMemory","type":"gauge","value":33518669824},{"id":"FreeMemory","type":"gauge","value":3527917568},{"id":"CPUutilization1","type":"gauge","value":59.793814433156726},{"id":"CPUutilization2","type":"gauge","value":46.487603307343086},{"id":"CPUutilization3","type":"gauge","value":42.47422680367953},{"id":"CPUutilization4","type":"gauge","value":25.63559321940101},{"id":"PollCount","type":"counter","delta":62}]`

	c, _ := setup(t)
	err := c.SetBatchMetrics(ctx, []byte(batch))
	assert.NoError(t, err)
}

func TestCollector_GetMetric(t *testing.T) {
	ctx := context.Background()
	c, _ := setup(t)
	_, err := c.GetMetric(ctx, "gauge", "Alloc")
	assert.Error(t, err)

	testVal := 123.456
	err = c.SetGaugeMetric(ctx, "Alloc", testVal)
	m, err := c.GetMetric(ctx, "gauge", "Alloc")
	assert.NoError(t, err)
	assert.Equal(t, "123.456", m)

	err = c.SetCounterMetric(ctx, constants.PollCount, 123)
	m, err = c.GetMetric(ctx, "counter", constants.PollCount)
	assert.NoError(t, err)
	assert.Equal(t, "123", m)

	m, err = c.GetMetric(ctx, "counterBad", constants.PollCount)
	assert.Error(t, err)
}

func TestCollector_GetAll(t *testing.T) {
	ctx := context.Background()
	c, _ := setup(t)

	testVal := 123.456
	c.SetGaugeMetric(ctx, "Alloc", testVal)
	c.SetCounterMetric(ctx, constants.PollCount, 123)

	all, err := c.GetAll(ctx)
	assert.NoError(t, err)

	assert.Equal(t, "Alloc: 123.456000\nPollCount: 123\n", all)

}

func TestCollector_Dump(t *testing.T) {
	c, _ := setup(t)

	err := c.GenerateDump()
	assert.NoError(t, err)

	err = c.LoadFromDump()
	assert.NoError(t, err)

}

func setup(t *testing.T) (*Collector, error) {

	cfg := &config.ServerConfig{}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	backupStorage := mock_collector.NewMockBackupStorage(ctrl)

	backupStorage.EXPECT().Save(`{"gauges":{},"counters":{}}`).Return(nil).AnyTimes()
	backupStorage.EXPECT().Load().Return(`{"gauges":{},"counters":{}}`, nil).AnyTimes()

	var (
		repo    ServerStorage
		collect *Collector
	)

	repo = storage.NewMemStorage()

	collect, err := NewCollector(cfg, repo, backupStorage)
	if err != nil {
		return nil, err
	}

	return collect, nil
}
