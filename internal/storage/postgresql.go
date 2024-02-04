package storage

import (
	"database/sql"
	"github.com/dnsoftware/go-metrics/internal/logger"

	//"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PgStorage struct {
	db *sql.DB
}

func NewPostgresqlStorage(dsn string) (*PgStorage, error) {

	// urlExample := "postgres://username:password@localhost:5432/database_name"
	//urlExample := "postgres://p1pool:Rextra516255@localhost:54321/metrics"
	////conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	//conn, err := pgx.Connect(context.Background(), urlExample)
	//if err != nil {
	//	fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
	//	os.Exit(1)
	//}
	//fmt.Println(conn)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Log().Error(err.Error())
		return nil, err
	}

	ps := &PgStorage{
		db: db,
	}

	return ps, nil
}

// Ping проверка работоспособности соединения с БД
func (p *PgStorage) Ping() bool {
	if p.db.Ping() != nil {
		return false
	}

	return true
}
