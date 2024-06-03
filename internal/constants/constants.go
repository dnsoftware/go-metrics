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
	ServerDefault     string = "localhost:8080"       // адрес:порт сервера по умолчанию
	StoreInterval     int64  = 300                    // интервалс сохранения значений метрик в файл
	FileStoragePath   string = "/tmp/metrics-db.json" // полное имя файла, куда сохраняются значения метрик
	RestoreSaved      bool   = true                   // загружать или нет ранее сохранённые значения из указанного файла при старте сервера
	AgentPprofAddr    string = ":8082"                // адрес:порт агента для работы с профилировщиком
	CryptoPublicFile  string = "certificate.pem"      // название файла с публичным ключом шифрования
	CryptoPrivateFile string = "privatekey.pem"       // название файла с приватным ключом шифрования
	//CryptoPublicFilePath  string = "/home/dmitry/go/src/go-metrics/internal/crypto/" + CryptoPublicFile
	//CryptoPrivateFilePath string = "/home/dmitry/go/src/go-metrics/internal/crypto/" + CryptoPrivateFile
	CryptoPublicFilePath  string = ""
	CryptoPrivateFilePath string = ""
	TrustedSubnet         string = ""
	GRPCDefault           string = "127.0.0.1:8090" // адрес:порт gRRC сервера по умолчанию
	ServerApi             string = "http"           // по какому протоколу клиент будет общаться с сервером (http || grpc) (флаг запуска -server-api, переменная окружения SERVER_API)
)

// Логгер.
const (
	LogFile  string = "./log.log"
	LogLevel        = zapcore.InfoLevel
)

// Типы контента.
const (
	TextPlain       string = "text/plain"
	TextHTML        string = "text/html"
	ApplicationJSON string = "application/json"
	ServerApiHTTP   string = "http"
	ServerApiGRPC   string = "grpc"
)

// Encoding
const (
	EncodingGzip      string = "gzip"
	CryptoHeaderName  string = "X-Content-Encoding" // ключ HTTP заголовка для асимметричного шифрования
	CryptoHeaderValue string = "crypto"             // значение HTTP заголовка CryptoHeaderName для асимметричного шифрования
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
const XRealIPName string = "X-Real-IP"

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

// Сообщения о завершении программы
const (
	MetricsUpdateCompleted                string = "Обновление метрик завершено..."
	MetricsGopsutilsUpdateCompletedstring string = "Обновление метрик gopcutils завершено..."
	PackagesPrepareCompleted              string = "Подготовка пакетов на отправку завершена..."
	SendMetricsCompleted                  string = "Отправка метрик завершена..."
	ProgramCompleted                      string = "Программа завершена!"
)
