package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dnsoftware/go-metrics/internal/constants"
	"github.com/dnsoftware/go-metrics/internal/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// PgStorage работает с Postgresql базой данных
type PgStorage struct {
	db *sql.DB
}

// Gauge для получения метрики counter из БД
type Gauge struct {
	name  string
	value float64
}

// Counter для получения метрики counter из БД
type Counter struct {
	name  string
	value int64
}

// DumpData карты gauges и counters для получения дампа БД
type DumpData struct {
	Gauges   map[string]float64 `json:"gauges"`
	Counters map[string]int64   `json:"counters"`
}

func NewPostgresqlStorage(dsn string) (*PgStorage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.DBContextTimeout)
	defer cancel()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Log().Error(err.Error())
		return nil, err
	}

	ps := &PgStorage{
		db: db,
	}

	// создание таблиц, если не существуют
	err = ps.CreateDatabaseTables(ctx)
	if err != nil {
		return nil, err
	}

	return ps, nil
}

// CreateDatabaseTables формирование структуры БД
func (p *PgStorage) CreateDatabaseTables(ctx context.Context) error {
	var query string

	// gauges
	query = `CREATE TABLE IF NOT EXISTS gauges
			(
			    id character varying(64) PRIMARY KEY,
			    val double precision NOT NULL,
			    updated_at timestamp with time zone NOT NULL
			)`

	err := p.retryExec(ctx, query)
	if err != nil {
		return err
	}

	// counters
	query = `CREATE TABLE IF NOT EXISTS counters
			(
			    id character varying(64) PRIMARY KEY,
			    val bigint NOT NULL,
			    updated_at timestamp with time zone NOT NULL
			)`

	err = p.retryExec(ctx, query)
	if err != nil {
		return err
	}

	return nil
}

// DropDatabaseTables удаление таблиц из базы
func (p *PgStorage) DropDatabaseTables(ctx context.Context) error {
	var query string

	// gauges
	query = `DROP TABLE gauges`

	err := p.retryExec(ctx, query)
	if err != nil {
		return err
	}

	// counters
	query = `DROP TABLE counters`

	err = p.retryExec(ctx, query)
	if err != nil {
		return err
	}

	return nil
}

// ClearDatabaseTables очистка таблиц базы.
func (p *PgStorage) ClearDatabaseTables(ctx context.Context) error {
	var query string

	// gauges
	query = `TRUNCATE TABLE gauges`

	err := p.retryExec(ctx, query)
	if err != nil {
		return err
	}

	// counters
	query = `TRUNCATE TABLE counters`

	err = p.retryExec(ctx, query)
	if err != nil {
		return err
	}

	return nil
}

// retryExec выполнение операции вставки/обновления несколькими попытками, если необходимо
func (p *PgStorage) retryExec(ctx context.Context, query string, args ...any) error {
	durations := strings.Split(constants.HTTPAttemtPeriods, ",")

	_, err := p.db.ExecContext(ctx, query, args...)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
		for _, duration := range durations {
			d, _ := time.ParseDuration(duration)
			time.Sleep(d)

			_, err = p.db.ExecContext(ctx, query, args...)
			if err == nil {
				break
			}
		}

		if err != nil {
			return fmt.Errorf("retryExec | ConnectionException: %w", err)
		}
	}

	if err != nil {
		return fmt.Errorf("retryExec: %w", err)
	}

	return nil
}

// SetGauge сохранение метрики типа gauge в хранилище.
// Параметры: name - название метрики, value - ее значение.
func (p *PgStorage) SetGauge(ctx context.Context, name string, value float64) error {
	query := `INSERT INTO gauges (id, val, updated_at)
			VALUES ($1, $2, now())
			ON CONFLICT (id)
			DO UPDATE
			SET id = $1, val = $2`

	err := p.retryExec(ctx, query, name, value)
	if err != nil {
		return fmt.Errorf("PgStorage | SetGauge: %w", err)
	}

	return nil
}

// GetGauge получение значения метрики типа gauge из хранилища.
// Параметры: name - название метрики.
func (p *PgStorage) GetGauge(ctx context.Context, name string) (float64, error) {
	query := `SELECT val FROM gauges WHERE id = $1`
	row := p.db.QueryRowContext(ctx, query, name)

	var val float64

	err := row.Scan(&val)
	if err != nil {
		return 0, fmt.Errorf("PgStorage | GetGauge: %w", err)
	}

	return val, nil
}

// SetCounter сохранение метрики типа counter в хранилище.
// Параметры: name - название метрики, value - ее значение.
func (p *PgStorage) SetCounter(ctx context.Context, name string, value int64) error {
	query := `INSERT INTO counters (id, val, updated_at)
			VALUES ($1, $2, now())
			ON CONFLICT (id)
			DO UPDATE
			SET id = $1, val = $2`

	err := p.retryExec(ctx, query, name, value)
	if err != nil {
		return fmt.Errorf("PgStorage | SetCounter: %w", err)
	}

	return nil
}

// GetCounter получение значения метрики типа counter из хранилища.
// Параметры: name - название метрики.
func (p *PgStorage) GetCounter(ctx context.Context, name string) (int64, error) {
	query := `SELECT val FROM counters WHERE id = $1`
	row := p.db.QueryRowContext(ctx, query, name)

	var val int64

	err := row.Scan(&val)
	if err != nil {
		return 0, fmt.Errorf("PgStorage | GetCounter: %w", err)
	}

	return val, nil
}

// SetBatch сохраняет метрики в базу пакетом из нескольких штук
func (p *PgStorage) SetBatch(ctx context.Context, batch []byte) error {
	var metrics []Metrics

	err := json.Unmarshal(batch, &metrics)
	if err != nil {
		return fmt.Errorf("PgStorage | SetBatch | json.Unmarshal: %w", err)
	}

	/* Логика нижеследующего кода (реализация одного запроса INSERT со множеством значений сразу):

	Используется SQL запрос вида:
		INSERT INTO gauges (id, val, updated_at)
		VALUES ($1, $2, now()), ($3, $4, now()),...
		ON CONFLICT (id)
		DO UPDATE
		SET val = EXCLUDED.val, updated_at = now()

	параметры для подстановки имеют нумерацию по принципу:
		$1, $2 - название метрики, значение метрики для первой записи
		$3, $4 - название метрики, значение метрики для второй записи
		...

	поэтому формируем срезы таких данных:
		gaugesKeyVal - для gauges
		countersKeyVal - для counters
	для дальнейшей генерации SQL запроса:
		p.db.ExecContext(ctx, query, gaugesKeyVal...)
		p.db.ExecContext(ctx, query, countersKeyVal...)

	далее вылез еще один момент:
		при обновлении в одном пакете одной и той же записи в базе, код
			ON CONFLICT (id)
			DO UPDATE
			SET val = EXCLUDED.val, updated_at = now()
		дает ошибку
			ERROR: ON CONFLICT DO UPDATE command cannot affect row a second time (SQLSTATE 21000)
		то есть второй раз не может повлиять на запись
		из-за этого не проходит автотест, с данными вида:
			[{"id":"CounterBatchZip146","type":"counter","delta":35154714},
			 {"id":"GaugeBatchZip188","type":"gauge","value":18032.255593532198},
			 {"id":"CounterBatchZip146","type":"counter","delta":1872525169},
			 {"id":"GaugeBatchZip188","type":"gauge","value":37453.22976261069}]

	поэтому предварительно формируется карты метрик (data), куда записываются:
		- последнее значение gauge метрики, так как она перезаписывает текущую
		- сумма значений counter метрик, так как они прибавляются к текущему
		что дает нам однократное обновление уникальных записей в запросе

	может быть есть более изящное решение, но я его не придумал))
	доклад закончил!)))
	*/

	// карта предварительно подготовленных метрик
	data := make(map[string]Metrics)

	for _, mt := range metrics {
		if mt.MType == constants.Gauge {
			data[mt.ID] = mt
		}

		if mt.MType == constants.Counter {
			if v, ok := data[mt.ID]; ok {
				vd := *v.Delta + *mt.Delta
				v.Delta = &vd
				data[mt.ID] = v

				continue
			}

			data[mt.ID] = mt
		}
	}

	var (
		gaugesKeyVal     []any    // срез пар значений для подстановки в SQL запрос вставки/обновления gauges
		countersKeyVal   []any    // срез пар значений для подстановки в SQL запрос вставки/обновления counters
		gaugeTemplates   []string // срез для формирования фрагмента множественной вставки gauges
		counterTemplates []string // срез для формирования фрагмента множественной вставки counters
		g                int64    // счетчик цикла gauges
		c                int64    // счетчик цикла counters
	)

	// формирование данных для генерации запроса множественной вставки
	for _, mt := range data {
		if mt.MType == constants.Gauge {
			g++
			val1 := g
			g++
			val2 := g
			gaugeTemplates = append(gaugeTemplates, fmt.Sprintf("($%d, $%d, now())", val1, val2))
			gaugesKeyVal = append(gaugesKeyVal, mt.ID, mt.Value)
		}

		if mt.MType == constants.Counter {
			c++
			val1 := c
			c++
			val2 := c
			counterTemplates = append(counterTemplates, fmt.Sprintf("($%d, $%d, now())", val1, val2))
			countersKeyVal = append(countersKeyVal, mt.ID, mt.Delta)
		}
	}

	// старт транзакции
	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("PgStorage | SetBatch | p.db.Begin(): %w", err)
	}

	if len(gaugeTemplates) > 0 {
		query := `INSERT INTO gauges (id, val, updated_at)
			VALUES ` + strings.Join(gaugeTemplates, ",") + `
			ON CONFLICT (id)
			DO UPDATE
			SET val = EXCLUDED.val, updated_at = now()`

		errR := p.retryExec(ctx, query, gaugesKeyVal...)
		if errR != nil {
			tx.Rollback()
			return fmt.Errorf("PgStorage | SetBatch | Upsert gauge: %w", errR)
		}
	}

	if len(counterTemplates) > 0 {
		query := `INSERT INTO counters (id, val, updated_at)
			VALUES ` + strings.Join(counterTemplates, ",") + `
			ON CONFLICT (id)
			DO UPDATE
			SET val = counters.val + EXCLUDED.val, updated_at = now()`

		errR := p.retryExec(ctx, query, countersKeyVal...)
		if errR != nil {
			tx.Rollback()
			return fmt.Errorf("PgStorage | SetBatch | Upsert counter: %w", errR)
		}
	}

	// завершаем транзакцию
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("PgStorage | SetBatch | Commit: %w", err)
	}

	return nil
}

// GetAll возврат всех метрик (карт gauge и counters)
func (p *PgStorage) GetAll(ctx context.Context) (map[string]float64, map[string]int64, error) {
	// gauges
	gRows, err := p.db.QueryContext(ctx, `SELECT id, val FROM gauges`)
	if err != nil {
		return nil, nil, fmt.Errorf("PgStorage | GetAll | gauges: %w", err)
	}
	defer gRows.Close()

	dump := DumpData{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}

	for gRows.Next() {
		v := Gauge{}

		err = gRows.Scan(&v.name, &v.value)
		if err != nil {
			return nil, nil, fmt.Errorf("PgStorage | GetAll | gauges Next: %w", err)
		}

		dump.Gauges[v.name] = v.value
	}

	err = gRows.Err()
	if err != nil {
		return nil, nil, fmt.Errorf("PgStorage | GetAll | gauges Next during iteration: %w", err)
	}

	// counters
	cRows, err := p.db.QueryContext(ctx, `SELECT id, val FROM counters`)
	if err != nil {
		return nil, nil, fmt.Errorf("PgStorage | GetAll | counters: %w", err)
	}
	defer cRows.Close()

	for cRows.Next() {
		v := Counter{}

		err = cRows.Scan(&v.name, &v.value)
		if err != nil {
			return nil, nil, fmt.Errorf("PgStorage | GetAll | counters Next: %w", err)
		}

		dump.Counters[v.name] = v.value
	}

	err = cRows.Err()
	if err != nil {
		return nil, nil, fmt.Errorf("PgStorage | GetAll | counters Next during iteration: %w", err)
	}

	return dump.Gauges, dump.Counters, nil
}

// GetDump получение json дампа БД
func (p *PgStorage) GetDump(ctx context.Context) (string, error) {
	dump := DumpData{}

	var err error

	dump.Gauges, dump.Counters, err = p.GetAll(ctx)
	if err != nil {
		return "", fmt.Errorf("PgStorage | GetDump | GetAll: %w", err)
	}

	data, err := json.Marshal(dump)
	if err != nil {
		return "", fmt.Errorf("PgStorage | GetDump | json.Marshal: %w", err)
	}

	return string(data), nil
}

// RestoreFromDump восстановление БД из json дампа
func (p *PgStorage) RestoreFromDump(ctx context.Context, dump string) error {
	data := DumpData{}

	err := json.Unmarshal([]byte(dump), &data)
	if err != nil {
		return fmt.Errorf("PgStorage | RestoreFromDump | json.Unmarshal: %w", err)
	}

	// старт транзакции
	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("PgStorage | RestoreFromDump | Tx begin: %w", err)
	}

	queryDel := `TRUNCATE gauges, counters`

	err = p.retryExec(ctx, queryDel)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("PgStorage | RestoreFromDump | Truncate: %w", err)
	}

	stmt, err := p.db.Prepare(`INSERT INTO gauges (id, val, updated_at)
			VALUES ($1, $2, now())`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("PgStorage | RestoreFromDump | Prepare gauges: %w", err)
	}
	defer stmt.Close()

	for name, val := range data.Gauges {
		_, err = stmt.ExecContext(ctx, name, val)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("PgStorage | RestoreFromDump | Insert gauge: %w", err)
		}
	}

	stmt, err = p.db.Prepare(`INSERT INTO counters (id, val, updated_at)
			VALUES ($1, $2, now())`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("PgStorage | RestoreFromDump | Prepare counters: %w", err)
	}
	defer stmt.Close()

	for name, val := range data.Counters {
		_, err = stmt.ExecContext(ctx, name, val)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("PgStorage | RestoreFromDump | Insert counter: %w", err)
		}
	}

	// завершаем транзакцию
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("PgStorage | RestoreFromDump | Commit: %w", err)
	}

	return nil
}

// DatabasePing проверка работоспособности соединения с БД
func (p *PgStorage) DatabasePing(ctx context.Context) bool {
	return p.db.PingContext(ctx) == nil
}
