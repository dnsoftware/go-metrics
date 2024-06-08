package collector

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"

	mock_collector "github.com/dnsoftware/go-metrics/internal/server/collector/mocks"
	"github.com/dnsoftware/go-metrics/internal/server/config"
	"github.com/dnsoftware/go-metrics/internal/storage"
)

func BenchmarkGetAll(b *testing.B) {
	b.StopTimer()

	ctx := context.Background()
	cfg := &config.ServerConfig{}

	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	backupStorage := mock_collector.NewMockBackupStorage(ctrl)

	backupStorage.EXPECT().Load().Return("", nil).AnyTimes()

	var (
		repo    ServerStorage
		collect *Collector
	)

	repo = storage.NewMemStorage()

	collect, err := NewCollector(cfg, repo, backupStorage)
	if err != nil {
		panic(err)
	}

	batch := `[{"id":"Alloc","type":"gauge","value":343728},{"id":"BuckHashSys","type":"gauge","value":7321},{"id":"Frees","type":"gauge","value":268},{"id":"GCCPUFraction","type":"gauge","value":0},{"id":"GCSys","type":"gauge","value":1758064},{"id":"HeapAlloc","type":"gauge","value":343728},{"id":"HeapIdle","type":"gauge","value":2490368},{"id":"HeapInuse","type":"gauge","value":1179648},{"id":"HeapObjects","type":"gauge","value":1959},{"id":"HeapReleased","type":"gauge","value":2490368},{"id":"HeapSys","type":"gauge","value":3670016},{"id":"LastGC","type":"gauge","value":0},{"id":"Lookups","type":"gauge","value":0},{"id":"MCacheInuse","type":"gauge","value":4800},{"id":"MCacheSys","type":"gauge","value":15600},{"id":"MSpanInuse","type":"gauge","value":54400},{"id":"MSpanSys","type":"gauge","value":65280},{"id":"Mallocs","type":"gauge","value":2227},{"id":"NextGC","type":"gauge","value":4194304},{"id":"NumForcedGC","type":"gauge","value":0},{"id":"NumGC","type":"gauge","value":0},{"id":"OtherSys","type":"gauge","value":1126423},{"id":"PauseTotalNs","type":"gauge","value":0},{"id":"StackInuse","type":"gauge","value":524288},{"id":"StackSys","type":"gauge","value":524288},{"id":"Sys","type":"gauge","value":7166992},{"id":"TotalAlloc","type":"gauge","value":343728},{"id":"RandomValue","type":"gauge","value":0.5116380300334399},{"id":"TotalMemory","type":"gauge","value":33518669824},{"id":"FreeMemory","type":"gauge","value":3527917568},{"id":"CPUutilization1","type":"gauge","value":59.793814433156726},{"id":"CPUutilization2","type":"gauge","value":46.487603307343086},{"id":"CPUutilization3","type":"gauge","value":42.47422680367953},{"id":"CPUutilization4","type":"gauge","value":25.63559321940101},{"id":"PollCount","type":"counter","delta":62}]`
	collect.storage.SetBatch(ctx, []byte(batch))

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		data, err := collect.GetAll(ctx)
		if err != nil {
			panic(err)
		}
		_ = data
	}

}
