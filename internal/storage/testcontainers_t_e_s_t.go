package storage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// образец тестов Postgresql в контейнере. Нужны для интеграционного тестирования.
func TestContainers(t *testing.T) {
	ctx := context.Background()

	dbname := "users"
	user := "user"
	password := "password"

	// 1. Start the postgres container
	container, err := postgres.RunContainer(
		ctx,
		testcontainers.WithImage("docker.io/postgres:16-alpine"),
		postgres.WithDatabase(dbname),
		postgres.WithUsername(user),
		postgres.WithPassword(password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Run any migrations on the database
	_, _, err = container.Exec(ctx, []string{"psql", "-U", user, "-d", dbname, "-c", "CREATE TABLE users (id SERIAL, name TEXT NOT NULL, age INT NOT NULL)"})
	if err != nil {
		t.Fatal(err)
	}

	// 2. Create a snapshot of the database to restore later
	err = container.Snapshot(ctx, postgres.WithSnapshotName("test-snapshot"))
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	dbURL, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Test inserting a user", func(t *testing.T) {
		t.Cleanup(func() {
			// 3. In each test, reset the DB to its snapshot state.
			err = container.Restore(ctx)
			if err != nil {
				t.Fatal(err)
			}
		})

		conn, err := pgx.Connect(context.Background(), dbURL)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close(context.Background())

		_, err = conn.Exec(ctx, "INSERT INTO users(name, age) VALUES ($1, $2)", "test", 42)
		if err != nil {
			t.Fatal(err)
		}

		var name string
		var age int64
		err = conn.QueryRow(context.Background(), "SELECT name, age FROM users LIMIT 1").Scan(&name, &age)
		if err != nil {
			t.Fatal(err)
		}

		if name != "test" {
			t.Fatalf("Expected %s to equal `test`", name)
		}
		if age != 42 {
			t.Fatalf("Expected %d to equal `42`", age)
		}
	})

	// 4. Run as many tests as you need, they will each get a clean database
	t.Run("Test querying empty DB", func(t *testing.T) {
		t.Cleanup(func() {
			err = container.Restore(ctx)
			if err != nil {
				t.Fatal(err)
			}
		})

		conn, err := pgx.Connect(context.Background(), dbURL)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close(context.Background())

		var name string
		var age int64
		err = conn.QueryRow(context.Background(), "SELECT name, age FROM users LIMIT 1").Scan(&name, &age)
		if !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("Expected error to be a NoRows error, since the DB should be empty on every test. Got %s instead", err)
		}
	})
}
