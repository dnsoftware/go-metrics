package storage

import (
	"database/sql"
	"fmt"
	//"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PgStorage struct {
}

func NewPostgresqlStorage() (PgStorage, error) {

	//urlExample := "postgres://p1pool:Rextra516255@localhost:54321/metrics"
	////conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	//conn, err := pgx.Connect(context.Background(), urlExample)
	//if err != nil {
	//	fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
	//	os.Exit(1)
	//}
	//fmt.Println(conn)

	auth := fmt.Sprintf("host=%s user=%s password=%s port=%s dbname=%s sslmode=disable",
		`localhost`, `p1pool`, `Rextra516255`, `54321`, `metrics`)
	db, err := sql.Open("pgx", auth)
	if err != nil {
		panic(err)
	}
	fmt.Println(db)

	ps := PgStorage{}

	return ps, nil
}
