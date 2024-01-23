package constants

import "go.uber.org/zap/zapcore"

// интервалы
const (
	PollInterval   int64 = 2  // интервал обновления метрик
	ReportInterval int64 = 10 // интервал отправки метрик на сервер
)

// действия
const (
	UpdateAction string = "update" // сохранить метрику
	ValueAction  string = "value"  // получить метрику
)

// типы метрик
const (
	Gauge   string = "gauge"
	Counter string = "counter"
)

// названия параметров
const (
	MetricType  string = "metricType"
	MetricName  string = "metricName"
	MetricValue string = "metricValue"
)

// метрики
const (
	PollCount   string = "PollCount"   // имя метрики счетчика
	RandomValue string = "RandomValue" // имя случайной метрики
)

const (
	ServerDefault string = "localhost:8080" // адрес:порт сервера по умолчанию
)

// логгер
const (
	LogFile  string = "./log.log"
	LogLevel        = zapcore.InfoLevel
)
