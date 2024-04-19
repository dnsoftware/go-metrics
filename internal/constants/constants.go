package constants

import (
	"go.uber.org/zap/zapcore"
	"time"
)

// интервалы
const (
	PollInterval       int64         = 2                               // интервал обновления метрик
	ReportInterval     int64         = 10                              // интервал отправки метрик на сервер
	BackupPeriod       int64         = 300                             // период в секундах, при указании которого происходит сохранение метрик в файл
	BackupPeriodSync   int64         = 0                               // период в секундах, при указании которого происходит синхронное с получением данных сохранение в файл
	DBContextTimeout   time.Duration = time.Duration(5) * time.Second  // длительность запроса в контексте работы с БД
	HTTPContextTimeout time.Duration = time.Duration(10) * time.Second // длительность запроса в контексте работы с сетью
)

// действия
const (
	UpdateAction  string = "update"  // сохранить метрику
	ValueAction   string = "value"   // получить метрику
	UpdatesAction string = "updates" // получить список метрик
	PprofAction   string = "/debug/pprof/"
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

	RestoreSavedEnv string = "RESTORE"
)

// метрики
const (
	PollCount   string = "PollCount"   // имя метрики счетчика
	RandomValue string = "RandomValue" // имя случайной метрики
)

// параметры работы сервера по умолчанию
const (
	ServerDefault   string = "localhost:8080"       // адрес:порт сервера по умолчанию
	StoreInterval   int64  = 300                    // интервалс сохранения значений метрик в файл
	FileStoragePath string = "/tmp/metrics-db.json" // полное имя файла, куда сохраняются значения метрик
	RestoreSaved    bool   = true                   // загружать или нет ранее сохранённые значения из указанного файла при старте сервера
	AgentPprofAddr  string = ":8082"                // адрес:порт агента для работы с профилировщиком
)

// логгер
const (
	LogFile  string = "./log.log"
	LogLevel        = zapcore.InfoLevel
)

// тип контента
const (
	TextPlain       string = "text/plain"
	TextHTML        string = "text/html"
	ApplicationJSON string = "application/json"
)

// Encoding
const (
	EncodingGzip string = "gzip"
)

// тип хранилища данных
const (
	Memory string = "memory"
	DBMS   string = "DBMS"
)

// Периоды повтора для Retriable ошибок
const (
	HTTPAttemtPeriods string = "1s,2s,5s"
	DBAttemtPeriods   string = "1s,2s,5s"
)

// Имена заголовков
const HashHeaderName string = "HashSHA256"

// для gopcutils
const (
	TotalMemory            string = "TotalMemory"
	FreeMemory             string = "FreeMemory"
	CPUutilization         string = "CPUutilization"
	CPUIntervalUtilization int64  = 5 // в секундах
)

// for workerpool
const (
	RateLimit      int = 3 // количество одновременно исходящих запросов на сервер по умолчанию (кол-во воркеров)
	BatchItemCount int = 5 // кол-во метрик в пакете
	ChannelCap     int = 5 // емкость канала
)
