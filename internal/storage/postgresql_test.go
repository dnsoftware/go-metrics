package storage

import (
	"context"
	"encoding/json"
	"flag"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"github.com/dnsoftware/go-metrics/internal/constants"

	"github.com/dnsoftware/go-metrics/internal/server/config"
)

var cfg = &config.ServerConfig{}
var ddsn string

func init() {
	godotenv.Load("../../.env_test")
	_ = env.Parse(cfg)

	flag.StringVar(&ddsn, "d", "", "data source name")
}

func setup() (*PgStorage, error) {

	if ddsn == "" {
		ddsn = cfg.DatabaseDSN
	}

	if ddsn == "" {
		ddsn = constants.TestDSN
	}

	pgStorage, err := NewPostgresqlStorage(ddsn)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	err = pgStorage.CreateDatabaseTables(ctx)
	if err != nil {
		return nil, err
	}

	return pgStorage, nil
}

func TestNewPostgresqlStorage(t *testing.T) {
	_, err := setup()
	require.NoError(t, err)
}

func TestPostgresqlSetGetGauge(t *testing.T) {
	ctx := context.Background()
	pgs, _ := setup()

	var testVal float64 = 123.456
	err := pgs.SetGauge(ctx, "test245", testVal)
	assert.NoError(t, err)

	val, err := pgs.GetGauge(ctx, "test245")
	assert.NoError(t, err)
	assert.Equal(t, testVal, val)

	err = pgs.DropDatabaseTables(ctx)
	assert.NoError(t, err)

	err = pgs.SetGauge(ctx, "test245", testVal)
	assert.Error(t, err)

	val, err = pgs.GetGauge(ctx, "test245")
	assert.Error(t, err)

}

func TestPostgresqlSetGetCounter(t *testing.T) {
	ctx := context.Background()
	pgs, _ := setup()

	var testVal int64 = 123
	err := pgs.SetCounter(ctx, "test245", testVal)
	assert.NoError(t, err)

	val, err := pgs.GetCounter(ctx, "test245")
	assert.NoError(t, err)
	assert.Equal(t, testVal, val)

	err = pgs.DropDatabaseTables(ctx)
	assert.NoError(t, err)

	err = pgs.SetCounter(ctx, "test245", testVal)
	assert.Error(t, err)

	val, err = pgs.GetCounter(ctx, "test245")
	assert.Error(t, err)

}

func TestPostgresqlSetBatch(t *testing.T) {
	ctx := context.Background()
	pgs, _ := setup()

	err := pgs.ClearDatabaseTables(ctx)
	assert.NoError(t, err)

	batch := `[{"id":"Alloc","type":"gauge","value":343728},{"id":"BuckHashSys","type":"gauge","value":7321},{"id":"Frees","type":"gauge","value":268},{"id":"GCCPUFraction","type":"gauge","value":0},{"id":"GCSys","type":"gauge","value":1758064},{"id":"HeapAlloc","type":"gauge","value":343728},{"id":"HeapIdle","type":"gauge","value":2490368},{"id":"HeapInuse","type":"gauge","value":1179648},{"id":"HeapObjects","type":"gauge","value":1959},{"id":"HeapReleased","type":"gauge","value":2490368},{"id":"HeapSys","type":"gauge","value":3670016},{"id":"LastGC","type":"gauge","value":0},{"id":"Lookups","type":"gauge","value":0},{"id":"MCacheInuse","type":"gauge","value":4800},{"id":"MCacheSys","type":"gauge","value":15600},{"id":"MSpanInuse","type":"gauge","value":54400},{"id":"MSpanSys","type":"gauge","value":65280},{"id":"Mallocs","type":"gauge","value":2227},{"id":"NextGC","type":"gauge","value":4194304},{"id":"NumForcedGC","type":"gauge","value":0},{"id":"NumGC","type":"gauge","value":0},{"id":"OtherSys","type":"gauge","value":1126423},{"id":"PauseTotalNs","type":"gauge","value":0},{"id":"StackInuse","type":"gauge","value":524288},{"id":"StackSys","type":"gauge","value":524288},{"id":"Sys","type":"gauge","value":7166992},{"id":"TotalAlloc","type":"gauge","value":343728},{"id":"RandomValue","type":"gauge","value":0.5116380300334399},{"id":"TotalMemory","type":"gauge","value":33518669824},{"id":"FreeMemory","type":"gauge","value":3527917568},{"id":"CPUutilization1","type":"gauge","value":59.793814433156726},{"id":"CPUutilization2","type":"gauge","value":46.487603307343086},{"id":"CPUutilization3","type":"gauge","value":42.47422680367953},{"id":"CPUutilization4","type":"gauge","value":25.63559321940101},{"id":"PollCount","type":"counter","delta":62}]`

	err = pgs.SetBatch(ctx, []byte(batch))
	assert.NoError(t, err)

	// неправильный формат json
	batch = `[{"id":"Alloc","type":`
	err = pgs.SetBatch(ctx, []byte(batch))
	var e *json.SyntaxError
	assert.ErrorAs(t, err, &e)

	// проверка правильности добавленных данных
	valGauge, err := pgs.GetGauge(ctx, "TotalAlloc")
	assert.Equal(t, float64(343728), valGauge)

	valCounter, err := pgs.GetCounter(ctx, "PollCount")
	assert.Equal(t, int64(62), valCounter)

}

func TestPostgresqlGetAll(t *testing.T) {
	ctx := context.Background()
	pgs, _ := setup()

	err := pgs.ClearDatabaseTables(ctx)
	assert.NoError(t, err)

	// пустая база
	gauges, counters, err := pgs.GetAll(ctx)
	assert.NoError(t, err)
	assert.Empty(t, gauges)
	assert.Empty(t, counters)

	// заполняем базу
	batch := `[{"id":"Alloc","type":"gauge","value":343728},{"id":"BuckHashSys","type":"gauge","value":7321},{"id":"Frees","type":"gauge","value":268},{"id":"GCCPUFraction","type":"gauge","value":0},{"id":"GCSys","type":"gauge","value":1758064},{"id":"HeapAlloc","type":"gauge","value":343728},{"id":"HeapIdle","type":"gauge","value":2490368},{"id":"HeapInuse","type":"gauge","value":1179648},{"id":"HeapObjects","type":"gauge","value":1959},{"id":"HeapReleased","type":"gauge","value":2490368},{"id":"HeapSys","type":"gauge","value":3670016},{"id":"LastGC","type":"gauge","value":0},{"id":"Lookups","type":"gauge","value":0},{"id":"MCacheInuse","type":"gauge","value":4800},{"id":"MCacheSys","type":"gauge","value":15600},{"id":"MSpanInuse","type":"gauge","value":54400},{"id":"MSpanSys","type":"gauge","value":65280},{"id":"Mallocs","type":"gauge","value":2227},{"id":"NextGC","type":"gauge","value":4194304},{"id":"NumForcedGC","type":"gauge","value":0},{"id":"NumGC","type":"gauge","value":0},{"id":"OtherSys","type":"gauge","value":1126423},{"id":"PauseTotalNs","type":"gauge","value":0},{"id":"StackInuse","type":"gauge","value":524288},{"id":"StackSys","type":"gauge","value":524288},{"id":"Sys","type":"gauge","value":7166992},{"id":"TotalAlloc","type":"gauge","value":343728},{"id":"RandomValue","type":"gauge","value":0.5116380300334399},{"id":"TotalMemory","type":"gauge","value":33518669824},{"id":"FreeMemory","type":"gauge","value":3527917568},{"id":"CPUutilization1","type":"gauge","value":59.793814433156726},{"id":"CPUutilization2","type":"gauge","value":46.487603307343086},{"id":"CPUutilization3","type":"gauge","value":42.47422680367953},{"id":"CPUutilization4","type":"gauge","value":25.63559321940101},{"id":"PollCount","type":"counter","delta":62}]`

	err = pgs.SetBatch(ctx, []byte(batch))
	assert.NoError(t, err)

	// проверка на заполненность
	gauges, counters, _ = pgs.GetAll(ctx)
	assert.NotEmpty(t, gauges)
	assert.NotEmpty(t, counters)

	// проверка первого и последнего добавленных значений
	valGauge, err := pgs.GetGauge(ctx, "Alloc")
	assert.Equal(t, float64(343728), valGauge)

	valCounter, err := pgs.GetCounter(ctx, "PollCount")
	assert.Equal(t, int64(62), valCounter)

	// проверка количества добавленных записей
	assert.Equal(t, 34, len(gauges))
	assert.Equal(t, 1, len(counters))

	// тестирование ошибок
	pgs.DropDatabaseTables(ctx)
	_, _, err = pgs.GetAll(ctx)
	assert.Error(t, err)

	var pgErr *pgconn.PgError
	assert.ErrorAs(t, err, &pgErr)

}

func TestPostgresqlGetDump(t *testing.T) {
	ctx := context.Background()
	pgs, _ := setup()

	pgs.ClearDatabaseTables(ctx)

	dump, err := pgs.GetDump(ctx)
	assert.NoError(t, err)
	assert.Equal(t, `{"gauges":{},"counters":{}}`, dump)

	var testVal float64 = 123.456
	pgs.SetGauge(ctx, "test245", testVal)
	dump, err = pgs.GetDump(ctx)
	assert.NoError(t, err)
	assert.Equal(t, `{"gauges":{"test245":123.456},"counters":{}}`, dump)

	var testCounter int64 = 123
	err = pgs.SetCounter(ctx, "test245", testCounter)
	dump, err = pgs.GetDump(ctx)
	assert.NoError(t, err)
	assert.Equal(t, `{"gauges":{"test245":123.456},"counters":{"test245":123}}`, dump)

}

func TestPostgresqlRestoreFromDump(t *testing.T) {
	ctx := context.Background()
	pgs, _ := setup()

	pgs.ClearDatabaseTables(ctx)

	dump := `{"gauges":{"test245":123.456},"counters":{"test245":123}}`
	err := pgs.RestoreFromDump(ctx, dump)
	assert.NoError(t, err)

	valGauge, err := pgs.GetGauge(ctx, "test245")
	assert.NoError(t, err)
	assert.Equal(t, 123.456, valGauge)

	valCounter, err := pgs.GetCounter(ctx, "test245")
	assert.NoError(t, err)
	assert.Equal(t, int64(123), valCounter)

}

func TestPostgresqlRestoreFromDumpNegative(t *testing.T) {
	ctx := context.Background()
	pgs, _ := setup()

	pgs.ClearDatabaseTables(ctx)

	// некорректный json
	dump := `{"gauges":{"test245":123.456},"counters":{"test245":123}`
	err := pgs.RestoreFromDump(ctx, dump)
	var e *json.SyntaxError
	assert.ErrorAs(t, err, &e)

	dump = `{"gauges":{"test245":123.5},"counters":{"test245":123.123}}`
	err = pgs.RestoreFromDump(ctx, dump)
	var eUnmarshal *json.UnmarshalTypeError
	assert.ErrorAs(t, err, &eUnmarshal)

}

func TestPostgresqlRDatabasePing(t *testing.T) {
	ctx := context.Background()
	pgs, _ := setup()

	ok := pgs.DatabasePing(ctx)
	assert.True(t, ok)

	pgs.db.Close()
	ok = pgs.DatabasePing(ctx)
	assert.False(t, ok)

}
