package storage

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Тестирование на реальной базе Postgresql
// используем Testcontainers: https://golang.testcontainers.org/modules/postgres/
func TestPostgres(t *testing.T) {

	ctx := context.Background()

	dbname := "users"
	user := "user"
	password := "password"

	// 1. Start the postgres container
	container, err := postgres.RunContainer(
		ctx,
		testcontainers.WithImage("docker.io/postgres:16-alpine"),
		postgres.WithDatabase(dbname),
		postgres.WithUsername(user),
		postgres.WithPassword(password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		t.Fatal(err)
	}

	dsn, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}

	pgs, err := NewPostgresqlStorage(dsn)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Test PostgresqlSetGetGauge", func(t *testing.T) {

		err = pgs.CreateDatabaseTables(ctx)
		assert.NoError(t, err)

		var testVal = 123.456
		err = pgs.SetGauge(ctx, "test245", testVal)
		assert.NoError(t, err)

		val, err2 := pgs.GetGauge(ctx, "test245")
		assert.NoError(t, err2)
		assert.Equal(t, testVal, val)

		err2 = pgs.DropDatabaseTables(ctx)
		assert.NoError(t, err2)

		err2 = pgs.SetGauge(ctx, "test245", testVal)
		assert.Error(t, err2)

		_, err2 = pgs.GetGauge(ctx, "test245")
		assert.Error(t, err2)

	})

	t.Run("Test PostgresqlSetGetCounter", func(t *testing.T) {
		err = pgs.CreateDatabaseTables(ctx)
		assert.NoError(t, err)

		var testVal int64 = 123
		err = pgs.SetCounter(ctx, "test245", testVal)
		assert.NoError(t, err)

		val, err2 := pgs.GetCounter(ctx, "test245")
		assert.NoError(t, err2)
		assert.Equal(t, testVal, val)

		err = pgs.DropDatabaseTables(ctx)
		assert.NoError(t, err)

		err = pgs.SetCounter(ctx, "test245", testVal)
		assert.Error(t, err)

		_, err = pgs.GetCounter(ctx, "test245")
		assert.Error(t, err)

	})

	t.Run("Test PostgresqlSetBatch", func(t *testing.T) {
		err = pgs.CreateDatabaseTables(ctx)
		assert.NoError(t, err)

		err = pgs.ClearDatabaseTables(ctx)
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
		valGauge, _ := pgs.GetGauge(ctx, "TotalAlloc")
		assert.Equal(t, float64(343728), valGauge)

		valCounter, _ := pgs.GetCounter(ctx, "PollCount")
		assert.Equal(t, int64(62), valCounter)
	})

	t.Run("Test PostgresqlGetAll", func(t *testing.T) {
		err = pgs.CreateDatabaseTables(ctx)
		assert.NoError(t, err)

		err = pgs.ClearDatabaseTables(ctx)
		assert.NoError(t, err)

		// пустая база
		gauges, counters, err2 := pgs.GetAll(ctx)
		assert.NoError(t, err2)
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
		valGauge, _ := pgs.GetGauge(ctx, "Alloc")
		assert.Equal(t, float64(343728), valGauge)

		valCounter, _ := pgs.GetCounter(ctx, "PollCount")
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

	})

	t.Run("Test PostgresqlGetDump", func(t *testing.T) {
		err = pgs.CreateDatabaseTables(ctx)
		assert.NoError(t, err)

		pgs.ClearDatabaseTables(ctx)

		dump, err2 := pgs.GetDump(ctx)
		assert.NoError(t, err2)
		assert.Equal(t, `{"gauges":{},"counters":{}}`, dump)

		var testVal = 123.456
		pgs.SetGauge(ctx, "test245", testVal)
		dump, err = pgs.GetDump(ctx)
		assert.NoError(t, err)
		assert.Equal(t, `{"gauges":{"test245":123.456},"counters":{}}`, dump)

		var testCounter int64 = 123
		_ = pgs.SetCounter(ctx, "test245", testCounter)
		dump, err = pgs.GetDump(ctx)
		assert.NoError(t, err)
		assert.Equal(t, `{"gauges":{"test245":123.456},"counters":{"test245":123}}`, dump)

	})

	t.Run("Test PostgresqlRestoreFromDump", func(t *testing.T) {
		err = pgs.CreateDatabaseTables(ctx)
		assert.NoError(t, err)

		pgs.ClearDatabaseTables(ctx)

		dump := `{"gauges":{"test245":123.456},"counters":{"test245":123}}`
		err2 := pgs.RestoreFromDump(ctx, dump)
		assert.NoError(t, err2)

		valGauge, err3 := pgs.GetGauge(ctx, "test245")
		assert.NoError(t, err3)
		assert.Equal(t, 123.456, valGauge)

		valCounter, err4 := pgs.GetCounter(ctx, "test245")
		assert.NoError(t, err4)
		assert.Equal(t, int64(123), valCounter)

	})

	t.Run("Test PostgresqlRestoreFromDumpNegative", func(t *testing.T) {
		err = pgs.CreateDatabaseTables(ctx)
		assert.NoError(t, err)

		pgs.ClearDatabaseTables(ctx)

		// некорректный json
		dump := `{"gauges":{"test245":123.456},"counters":{"test245":123}`
		err2 := pgs.RestoreFromDump(ctx, dump)
		var e *json.SyntaxError
		assert.ErrorAs(t, err2, &e)

		dump = `{"gauges":{"test245":123.5},"counters":{"test245":123.123}}`
		err = pgs.RestoreFromDump(ctx, dump)
		var eUnmarshal *json.UnmarshalTypeError
		assert.ErrorAs(t, err, &eUnmarshal)

	})

	t.Run("Test PostgresqlRDatabasePing", func(t *testing.T) {
		err = pgs.CreateDatabaseTables(ctx)
		assert.NoError(t, err)

		ok := pgs.DatabasePing(ctx)
		assert.True(t, ok)

		pgs.db.Close()
		ok = pgs.DatabasePing(ctx)
		assert.False(t, ok)

	})

	if err = container.Terminate(ctx); err != nil {
		t.Fatalf("failed to terminate container: %s", err)
	}

}
