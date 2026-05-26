package integration_tests

import (
	"context"
	"log"
	"os"
	"testing"

	"concept-tracker/db"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var testPool *pgxpool.Pool
var testCtr *postgres.PostgresContainer

func TestMain(m *testing.M) {

	ctx := context.Background()

	dbName := "users"
	dbUser := "postgres"
	dbPassword := "password"

	// start db container
	ctr, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		postgres.BasicWaitStrategies(),
		postgres.WithSQLDriver("pgx"),
	)
	if err != nil {
		log.Fatal(err)
	}

	testCtr = ctr

	connStr, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	d, err := iofs.New(db.Migrations, "migrations")
	if err != nil {
		log.Fatal(err)
	}

	mig, err := migrate.NewWithSourceInstance("iofs", d, connStr)
	if err != nil {
		log.Fatal(err)
	}

	// run migrations
	err = mig.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}

	sourceErr, dbErr := mig.Close()
	if sourceErr != nil {
		log.Fatal(sourceErr)
	}
	if dbErr != nil {
		log.Fatal(dbErr)
	}

	// take snapshot
	err = ctr.Snapshot(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// create pgxpool conn
	testPool, err = pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	if err = ctr.Terminate(ctx); err != nil {
		log.Fatal(err)
	}
	os.Exit(code)
}

func restoreDB(t *testing.T, ctx context.Context) {
	t.Helper()
	if err := testCtr.Restore(ctx); err != nil {
		t.Fatal(err)
	}

	testPool.Close()
	connStr, err := testCtr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	testPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatal(err)
	}
}
