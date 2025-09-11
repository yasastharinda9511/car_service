package database

import "database/sql"

type Database struct {
	db *sql.DB
}

func NewDatabase(driver string, dsn string) (*sql.DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}
