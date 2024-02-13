package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dnsoftware/go-metrics/internal/constants"
	"github.com/dnsoftware/go-metrics/internal/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"strings"
	"time"
)

type PgStorage struct {
	storageType string
	db          *sql.DB
}

type Gauge struct {
	name  string
	value float64
}

type Counter struct {
	name  string
	value int64
}

type DumpData struct {
	Gauges   map[string]float64 `json:"gauges"`
	Counters map[string]int64   `json:"counters"`
}

func NewPostgresqlStorage(dsn string) (*PgStorage, error) {

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Log().Error(err.Error())
		return nil, err
	}

	ps := &PgStorage{
		storageType: constants.DBMS,
		db:          db,
	}

	// создание таблиц, если не существуют
	err = ps.createDatabaseTables()
	if err != nil {
		return nil, err
	}

	return ps, nil
}

// формирование структуры БД
func (p *PgStorage) createDatabaseTables() error {

	var query string

	// gauges
	query = `CREATE TABLE IF NOT EXISTS gauges
			(
			    id character varying(64) PRIMARY KEY,
			    val double precision NOT NULL,
			    updated_at timestamp with time zone NOT NULL
			)`
	err := p.retryExec(query)
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
	err = p.retryExec(query)
	if err != nil {
		return err
	}

	return nil
}

func (p *PgStorage) Type() string {
	return p.storageType
}

// Health проверка работоспособности соединения с БД
func (p *PgStorage) Health() bool {
	return p.db.Ping() == nil
}

func (p *PgStorage) retryExec(query string, args ...any) error {
	durations := strings.Split(constants.HTTPAttemtPeriods, ",")

	_, err := p.db.Exec(query, args...)

	var pgErr *pgconn.PgError
	//if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
	if errors.As(err, &pgErr) && pgerrcode.IsSyntaxErrororAccessRuleViolation(pgErr.Code) {
		for _, duration := range durations {
			d, _ := time.ParseDuration(duration)
			time.Sleep(d)
			_, err = p.db.Exec(query, args...)
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

func (p *PgStorage) SetGauge(name string, value float64) error {

	query := `INSERT INTO gauges (id, val, updated_at)
			VALUES ($1, $2, now())
			ON CONFLICT (id)
			DO UPDATE
			SET id = $1, val = $2`

	err := p.retryExec(query, name, value)
	if err != nil {
		return fmt.Errorf("PgStorage | SetGauge: %w", err)
	}

	return nil
}

func (p *PgStorage) GetGauge(name string) (float64, error) {

	query := `SELECT val FROM gauges WHERE id = $1`
	row := p.db.QueryRow(query, name)

	var val float64
	err := row.Scan(&val)
	if err != nil {
		return 0, fmt.Errorf("PgStorage | GetGauge: %w", err)
	}

	return val, nil
}

func (p *PgStorage) SetCounter(name string, value int64) error {

	query := `INSERT INTO counters (id, val, updated_at)
			VALUES ($1, $2, now())
			ON CONFLICT (id)
			DO UPDATE
			SET id = $1, val = $2`
	err := p.retryExec(query, name, value)
	if err != nil {
		return fmt.Errorf("PgStorage | SetCounter: %w", err)
	}

	return nil
}

func (p *PgStorage) SetBatch(batch []byte) error {
	var metrics []Metrics

	err := json.Unmarshal(batch, &metrics)
	if err != nil {
		return fmt.Errorf("PgStorage | SetBatch | json.Unmarshal: %w", err)
	}

	// старт транзакции
	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("PgStorage | SetBatch | p.db.Begin(): %w", err)
	}
	for _, mt := range metrics {
		err := errors.New("")
		if mt.MType == constants.Gauge {
			query := `INSERT INTO gauges (id, val, updated_at)
			VALUES ($1, $2, now())
			ON CONFLICT (id)
			DO UPDATE
			SET id = $1, val = $2, updated_at = now()`
			err = p.retryExec(query, mt.ID, mt.Value)
		}

		if mt.MType == constants.Counter {
			query := `INSERT INTO counters (id, val, updated_at)
			VALUES ($1, $2, now())
			ON CONFLICT (id)
			DO UPDATE
			SET val = counters.val + $2, updated_at = now()`
			err = p.retryExec(query, mt.ID, mt.Delta)
		}

		if err != nil {
			tx.Rollback()
			return fmt.Errorf("PgStorage | SetBatch | Upsert metric: %w", err)
		}
	}
	// завершаем транзакцию
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("PgStorage | SetBatch | Commit: %w", err)
	}

	return nil
}

func (p *PgStorage) GetCounter(name string) (int64, error) {

	query := `SELECT val FROM counters WHERE id = $1`
	row := p.db.QueryRow(query, name)

	var val int64
	err := row.Scan(&val)
	if err != nil {
		return 0, fmt.Errorf("PgStorage | GetCounter: %w", err)
	}

	return val, nil
}

// возврат карт gauge и counters
func (p *PgStorage) GetAll() (map[string]float64, map[string]int64, error) {

	// gauges
	gRows, err := p.db.Query(`SELECT id, val FROM gauges`)
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
	cRows, err := p.db.Query(`SELECT id, val FROM counters`)
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

// получение json дампа
func (p *PgStorage) GetDump() (string, error) {

	dump := DumpData{}

	err := *new(error)
	dump.Gauges, dump.Counters, err = p.GetAll()
	if err != nil {
		return "", fmt.Errorf("PgStorage | GetDump | GetAll: %w", err)
	}

	data, err := json.Marshal(dump)
	if err != nil {
		return "", fmt.Errorf("PgStorage | GetDump | json.Marshal: %w", err)
	}

	return string(data), nil
}

// восстановление из json дампа old version
func (p *PgStorage) OldRestoreFromDump(dump string) error {

	data := DumpData{}

	err := json.Unmarshal([]byte(dump), &data)
	if err != nil {
		return fmt.Errorf("PgStorage | RestoreFromDump | json.Unmarshal: %w", err)
	}

	for name, val := range data.Gauges {
		err = p.SetGauge(name, val)
		if err != nil {
			logger.Log().Error(err.Error())
		}
	}

	for name, val := range data.Counters {
		err = p.SetCounter(name, val)
		if err != nil {
			logger.Log().Error(err.Error())
		}
	}

	return nil
}

// восстановление из json дампа
func (p *PgStorage) RestoreFromDump(dump string) error {

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
	err = p.retryExec(queryDel)
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
		_, err = stmt.Exec(name, val)
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
		_, err = stmt.Exec(name, val)
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
