// Package constants содержит все используемые в проекте константы
package constants

import (
	"time"

	"go.uber.org/zap/zapcore"
)

// Различные интервалы.
const (
	PollInterval       int64         = 2                               // интервал обновления метрик
	ReportInterval     int64         = 10                              // интервал отправки метрик на сервер
	BackupPeriod       int64         = 300                             // период в секундах, при указании которого происходит сохранение метрик в файл
	BackupPeriodSync   int64         = 0                               // период в секундах, при указании которого происходит синхронное с получением данных сохранение в файл
	DBContextTimeout   time.Duration = time.Duration(5) * time.Second  // длительность запроса в контексте работы с БД
	HTTPContextTimeout time.Duration = time.Duration(10) * time.Second // длительность запроса в контексте работы с сетью
)

// Действия. Используются для построения url.
const (
	UpdateAction  string = "update"  // сохранить метрику
	ValueAction   string = "value"   // получить метрику
	UpdatesAction string = "updates" // получить список метрик
	PprofAction   string = "/debug/pprof/"
)

// Типы метрик.
const (
	Gauge   string = "gauge"
	Counter string = "counter"
)

// Названия параметров.
const (
	MetricType  string = "metricType"
	MetricName  string = "metricName"
	MetricValue string = "metricValue"

	RestoreSavedEnv string = "RESTORE"
)

// Метрики.
const (
	PollCount   string = "PollCount"   // имя метрики счетчика
	RandomValue string = "RandomValue" // имя случайной метрики
)

// Параметры работы сервера (по умолчанию).
const (
	ServerDefault   string = "localhost:8080"       // адрес:порт сервера по умолчанию
	StoreInterval   int64  = 300                    // интервалс сохранения значений метрик в файл
	FileStoragePath string = "/tmp/metrics-db.json" // полное имя файла, куда сохраняются значения метрик
	RestoreSaved    bool   = true                   // загружать или нет ранее сохранённые значения из указанного файла при старте сервера
	AgentPprofAddr  string = ":8082"                // адрес:порт агента для работы с профилировщиком
)

// Логгер.
const (
	LogFile  string = "./log.log"
	LogLevel        = zapcore.ErrorLevel
)

// Типы контента.
const (
	TextPlain       string = "text/plain"
	TextHTML        string = "text/html"
	ApplicationJSON string = "application/json"
)

// Encoding
const (
	EncodingGzip string = "gzip"
)

// Тип хранилища данных.
const (
	Memory string = "memory"
	DBMS   string = "DBMS"
)

// Периоды повтора для Retriable ошибок.
const (
	HTTPAttemtPeriods string = "1s,2s,5s"
	DBAttemtPeriods   string = "1s,2s,5s"
)

// HashHeaderName Имена заголовков.
const HashHeaderName string = "HashSHA256"

// Для gopcutils.
const (
	TotalMemory            string = "TotalMemory"
	FreeMemory             string = "FreeMemory"
	CPUutilization         string = "CPUutilization"
	CPUIntervalUtilization int64  = 5 // в секундах
)

// Для пула воркеров
const (
	RateLimit      int = 3 // количество одновременно исходящих запросов на сервер по умолчанию (кол-во воркеров)
	BatchItemCount int = 5 // кол-во метрик в пакете
	ChannelCap     int = 5 // емкость канала
)

// Тестирование
const (
	TestDSN string = "postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable"
	//TestDSN string = "postgres://praktikum:praktikum@127.0.0.1:5532/praktikum?sslmode=disable"
)

// Сообщения о завершении программы
const (
	MetricsUpdateCompleted                string = "Обновление метрик завершено..."
	MetricsGopsutilsUpdateCompletedstring string = "Обновление метрик gopcutils завершено..."
	PackagesPrepareCompleted              string = "Подготовка пакетов на отправку завершена..."
	SendMetricsCompleted                  string = "Отправка метрик завершена..."
	ProgramCompleted                      string = "Программа завершена!"
)
