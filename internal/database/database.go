package database

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Database struct {
	db *pgxpool.Pool
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

	return &Database{db}, nil
}
