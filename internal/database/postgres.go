package database

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect(dbUrl string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dbUrl)

	if err != nil {
		return nil, fmt.Errorf("failed to open database, %w\n", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database, %w\n", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5) //fixme parse from config

	return db, nil
}
