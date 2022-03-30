package postgres

import (
	"UrlShort/config"
	"UrlShort/internal/utils"
	"context"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jackc/tern/migrate"
	"log"
	"time"
)

type Client interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

func NewClient(ctx context.Context, cfg config.Storage) (pool *pgxpool.Pool, err error) { //Пул req'ов, подключение возможно не сразу
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	err = utils.DoWithTries(func() error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		pool, err = pgxpool.Connect(ctx, dsn)
		if err != nil {
			return err
		}
		return nil

	}, 5, 5*time.Second)

	if err != nil {
		log.Fatal()
	}

	log.Printf("DB. Succsess for connectiong to DB, storage: postgresql://%s:%s@%s:%s/%s\n", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		log.Fatalf("Unable to acquire a database connection: %v\n", err)
	}
	migrateDatabase(conn.Conn())
	conn.Release()

	return pool, nil
}

func migrateDatabase(conn *pgx.Conn) {
	migrator, err := migrate.NewMigrator(context.Background(), conn, "schema_version")
	if err != nil {
		log.Fatalf("DB: Unable to create a migrator: %v\n", err)
	}
	err = migrator.LoadMigrations("./migrations")
	if err != nil {
		log.Fatalf("DB: Unable to load migrations: %v\n", err)
	}

	err = migrator.Migrate(context.Background())
	if err != nil {
		log.Fatalf("DB: Unable to migrate: %v\n", err)
	}

	ver, err := migrator.GetCurrentVersion(context.Background())
	if err != nil {
		log.Fatalf("DB: Unable to get current schema version: %v\n", err)
	}

	log.Printf("DB: Migration done. Current schema version: %v\n", ver)
}
