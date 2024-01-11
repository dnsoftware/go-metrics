package domain

import (
	"github.com/dnsoftware/go-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetrics_repo(t *testing.T) {

	// работа с репозиторием
	repo := storage.NewMemStorage()

	checkVal, err := repo.GetGauge("Alloc")
	if assert.Error(t, err, "При извлечении из репозитория несуществующего значения gauge должна быть ошибка") {
		assert.Equal(t, float64(0), checkVal, "Должен быть 0 при извлечении несуществующего значения gauge")
	}

	checkValCnt, err := repo.GetCounter("PollCount")
	if assert.Error(t, err, "При извлечении из репозитория несуществующего значения counter должна быть ошибка") {
		assert.Equal(t, int64(0), checkValCnt, "Должен быть 0 при извлечении несуществующего значения counter")
	}

	testVal := float64(123456)
	repo.SetGauge("Alloc", testVal)
	checkVal, err = repo.GetGauge("Alloc")
	assert.Equal(t, testVal, checkVal, "Добавленное в репозиторий должно быть равно извлеченному из репозитория")
	assert.NoError(t, err, "Добавление/обновление в репозиторий должно быть без ошибок")

	testValCnt := int64(23456)
	repo.SetCounter("PollCount", testValCnt)
	checkValCnt, err = repo.GetCounter("PollCount")
	assert.Equal(t, testValCnt, checkValCnt, "Добавленное в репозиторий counter должно быть равно извлеченному из репозитория")
	assert.NoError(t, err, "Добавление/обновление в репозиторий counter должно быть без ошибок")

	// допустимые метрики

	//tests := []struct {
	//	name string
	//}{
	//	{
	//		name: "first",
	//	},
	//}
	//
	//for _, tt := range tests {
	//	t.Run(tt.name, func(t *testing.T) {
	//
	//		//sender := mocks.NewMetricsSender(t)
	//		repo := storage.NewMemStorage()
	//		//repo := mocks.NewAgentStorage(t)
	//
	//		testVal := float64(123456)
	//		repo.SetGauge("Alloc", testVal)
	//		mVal, err := repo.GetGauge("Alloc")
	//
	//		assert.Equal(t, testVal, mVal)
	//		assert.Equal(t, nil, err)
	//
	//		//m := NewMetrics(repo, sender)
	//		//m.updateMetrics()
	//	})
	//}

}
