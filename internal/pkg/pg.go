package pkg

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"

	"time"
)

type DB struct {
	*bun.DB
}

func Dial(dsn string) (*DB, error) {
	if dsn == "" {
		return nil, errors.New("no postgres DSN provided")
	}
	db := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(1 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	bunDb := bun.NewDB(db, pgdialect.New())

	bunDb.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
		bundebug.FromEnv("BUNDEBUG"),
	))
	return &DB{bunDb}, nil
}
