package database

import "database/sql"

type Database struct {
	Db *sql.DB
}

func NewDatabase(driver string, dsn string) (*Database, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Database{Db: db}, nil
}

func (d *Database) Close() error {
	return d.Db.Close()
}
