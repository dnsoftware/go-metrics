package constants

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

// метрики
const (
	PollCount   string = "PollCount"   // имя метрики счетчика
	RandomValue string = "RandomValue" // имя случайной метрики
)

const (
	ServerDefault string = "localhost:8080" // адрес:порт сервера по умолчанию
)
