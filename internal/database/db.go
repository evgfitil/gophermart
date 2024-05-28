package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/evgfitil/gophermart.git/internal/logger"
)

const (
	driverName    = "pgx"
	migrationPath = "db/migrations"
)

type DBStorage struct {
	conn *sql.DB
}

func NewDBStorage(databaseDSN string) (*DBStorage, error) {
	var db DBStorage
	conn, err := sql.Open(driverName, databaseDSN)

	if err != nil {
		logger.Sugar.Fatalf("unable to connect to database: %v", err)
		return nil, err
	}

	m, err := migrate.New(fmt.Sprintf("file://%s", migrationPath), databaseDSN)
	if err != nil {
		logger.Sugar.Fatalf("error with migrations: %v", err)
	}
	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Sugar.Infoln("skipping migrations, no changes")
		} else {
			logger.Sugar.Fatalf("error applying migrations: %v", err)
			return nil, err
		}
	} else {
		logger.Sugar.Infoln("migrations applied")
	}

	db = DBStorage{conn: conn}
	return &db, nil
}

func (db *DBStorage) Close() error {
	return db.conn.Close()
}

func (db *DBStorage) Ping(ctx context.Context) error {
	err := db.conn.PingContext(ctx)
	if err != nil {
		logger.Sugar.Errorf("error connecting to database: %v", err)
		return err
	}
	return nil
}
