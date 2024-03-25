package database

import (
	"context"
	"embed"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"time"
)

type Database struct {
	db  *pgxpool.Pool
	dsn string
}

type DBExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

func checkConnection(ctx context.Context, db *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()

	if err := db.Ping(ctx); err != nil {
		return err
	}

	return nil
}

func New(ctx context.Context, dsn string) (*Database, error) {
	db, err := pgxpool.New(ctx, dsn)

	if err != nil {
		return nil, err
	}

	if err := checkConnection(ctx, db); err != nil {
		return nil, err
	}

	return &Database{db, dsn}, nil
}

//go:embed migrations/*
var migrationsFS embed.FS

func (d *Database) RunMigrations() error {
	driver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return err
	}

	migrations, err := migrate.NewWithSourceInstance("iofs", driver, d.dsn)

	if err != nil {
		return err
	}

	err = migrations.Up()

	if err != nil {
		if err.Error() == "no change" {
			log.Printf("No new migrations found")
			return nil
		}

		return err
	}

	return nil
}
