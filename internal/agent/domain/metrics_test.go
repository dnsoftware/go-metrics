package domain

import (
	"context"
	"github.com/dnsoftware/go-metrics/internal/constants"
	"github.com/dnsoftware/go-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
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
