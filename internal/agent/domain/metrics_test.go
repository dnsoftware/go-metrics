package domain

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"
	"syscall"
	"testing"
	"time"

	mock_domain "github.com/dnsoftware/go-metrics/internal/agent/domain/mocks"

	"github.com/golang/mock/gomock"

	"github.com/shirou/gopsutil/v3/cpu"

	"github.com/dnsoftware/go-metrics/internal/constants"
	"github.com/dnsoftware/go-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetrics_repo(t *testing.T) {
	ctx := context.Background()

	// работа с репозиторием
	repo := storage.NewMemStorage()

	checkVal, err := repo.GetGauge(ctx, "Alloc")
	if assert.Error(t, err, "При извлечении из репозитория несуществующего значения gauge должна быть ошибка") {
		assert.Equal(t, float64(0), checkVal, "Должен быть 0 при извлечении несуществующего значения gauge")
	}

	checkValCnt, err := repo.GetCounter(ctx, constants.PollCount)
	if assert.Error(t, err, "При извлечении из репозитория несуществующего значения counter должна быть ошибка") {
		assert.Equal(t, int64(0), checkValCnt, "Должен быть 0 при извлечении несуществующего значения counter")
	}

	testVal := float64(123456)
	repo.SetGauge(ctx, "Alloc", testVal)
	checkVal, err = repo.GetGauge(ctx, "Alloc")
	assert.InEpsilon(t, testVal, checkVal, 0.0001, "Добавленное в репозиторий должно быть равно извлеченному из репозитория")
	require.NoError(t, err, "Добавление/обновление в репозиторий должно быть без ошибок")

	testValCnt := int64(23456)
	repo.SetCounter(ctx, constants.PollCount, testValCnt)
	checkValCnt, err = repo.GetCounter(ctx, constants.PollCount)
	assert.Equal(t, testValCnt, checkValCnt, "Добавленное в репозиторий counter должно быть равно извлеченному из репозитория")
	require.NoError(t, err, "Добавление/обновление в репозиторий counter должно быть без ошибок")
}

func TestUpdateMetricsReflect(t *testing.T) {
	ctx := context.Background()
	metrics := updateMetricsSetup()
	metrics.UpdateMetricsReflect()

	for _, val := range gaugeMetricsList {
		_, err := metrics.storage.GetGauge(ctx, val)
		assert.NoError(t, err)
	}

}

func TestUpdateMetrics(t *testing.T) {
	ctx := context.Background()
	metrics := updateMetricsSetup()
	metrics.UpdateMetrics()

	for _, val := range gaugeMetricsList {
		_, err := metrics.storage.GetGauge(ctx, val)
		assert.NoError(t, err)
	}

}

func TestUpdateGopcMetrics(t *testing.T) {
	ctx := context.Background()
	metrics := updateMetricsSetup()
	metrics.UpdateGopcMetrics()

	_, err := metrics.storage.GetGauge(ctx, constants.TotalMemory)
	assert.NoError(t, err)
	_, err = metrics.storage.GetGauge(ctx, constants.FreeMemory)
	assert.NoError(t, err)
	cc, _ := cpu.Percent(time.Second*time.Duration(constants.CPUIntervalUtilization), true)
	for key := range cc {
		_, err = metrics.storage.GetGauge(ctx, constants.CPUutilization+strconv.Itoa(key+1))
		assert.NoError(t, err)
	}
}

func TestSendMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repository := storage.NewMemStorage()
	sender := mock_domain.MockMetricsSender{}
	gopcMetricsList := []string{constants.TotalMemory, constants.FreeMemory}
	cpuCount, _ := cpu.Counts(false)
	for i := 1; i <= cpuCount; i++ {
		gopcMetricsList = append(gopcMetricsList, constants.CPUutilization+strconv.Itoa(i))
	}

	metrics := NewMetrics(repository, &sender, &fl{})
	sendCount := metrics.sendMetrics()
	assert.Equal(t, 1, sendCount)
	metrics.UpdateMetrics()
	sendCount = metrics.sendMetrics()
	assert.Equal(t, 29, sendCount)

}

func TestIsGauge(t *testing.T) {
	metrics := updateMetricsSetup()

	isGauge := metrics.IsGauge("Alloc")
	assert.True(t, isGauge)

	isGauge = metrics.IsGauge("NoAlloc")
	assert.False(t, isGauge)
}

// TestStart проверяет старт и завершение горутин работы с метриками
// проверяется вывод сообщений о завершении горутин в консоль
func TestStart(t *testing.T) {
	metrics := updateMetricsSetup()
	var output bytes.Buffer
	metrics.SetMessageWriter(&output)
	go metrics.Start()
	time.Sleep(2 * time.Second)
	syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	time.Sleep(16 * time.Second)

	metrics.mutex.Lock()
	assert.Contains(t, output.String(), constants.MetricsUpdateCompleted)
	assert.Contains(t, output.String(), constants.MetricsGopsutilsUpdateCompletedstring)
	assert.Contains(t, output.String(), constants.PackagesPrepareCompleted)
	assert.Contains(t, output.String(), constants.SendMetricsCompleted)
	assert.Contains(t, output.String(), constants.ProgramCompleted)
	metrics.mutex.Unlock()
}

// TestSendMetricsBatch проверяет отправку пакетов в канал
func TestSendMetricsBatch(t *testing.T) {
	ctx := context.Background()
	jobsCh := make(chan []byte, constants.ChannelCap)
	metrics := updateMetricsSetup()
	metrics.UpdateMetrics()
	go metrics.sendMetricsBatch(ctx, jobsCh)
	time.Sleep(1 * time.Second)

	assert.Equal(t, constants.ChannelCap, len(jobsCh))

	var batch []MetricsItem
	batchByte := <-jobsCh
	err := json.Unmarshal(batchByte, &batch)
	assert.NoError(t, err)
	assert.Equal(t, batch[0].ID, "Alloc")
}

func TestWorker(t *testing.T) {
	ctx := context.Background()
	jobsCh := make(chan []byte, constants.ChannelCap)

	repository := storage.NewMemStorage()
	sender := mock_domain.MockMetricsSender{}
	gopcMetricsList := []string{constants.TotalMemory, constants.FreeMemory}
	cpuCount, _ := cpu.Counts(false)
	for i := 1; i <= cpuCount; i++ {
		gopcMetricsList = append(gopcMetricsList, constants.CPUutilization+strconv.Itoa(i))
	}

	metrics := NewMetrics(repository, &sender, &fl{})

	metrics.UpdateMetrics()
	go metrics.sendMetricsBatch(ctx, jobsCh)
	time.Sleep(1 * time.Second)

	rateLimitChan := make(chan struct{}, constants.RateLimit)
	batchByte := <-jobsCh
	rateLimitChan <- struct{}{}
	metrics.worker(batchByte, rateLimitChan)
	assert.Equal(t, 0, len(rateLimitChan)) // worker забрал структуру из канала
}
