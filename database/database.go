package database

import (
	"car_service/logger"
	"database/sql"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(driver string, dsn string) (*sql.DB, error) {
	logger.WithField("driver", driver).Info("Connecting to database")

	db, err := sql.Open(driver, dsn)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"driver": driver,
			"error":  err.Error(),
		}).Error("Failed to open database connection")
		return nil, err
	}

	logger.Debug("Pinging database to verify connection")
	if err := db.Ping(); err != nil {
		logger.WithFields(map[string]interface{}{
			"driver": driver,
			"error":  err.Error(),
		}).Error("Failed to ping database")
		return nil, err
	}

	logger.WithField("driver", driver).Info("Database connection established successfully")
	return db, nil
}

func (d *Database) Close() error {
	logger.Info("Closing database connection")
	err := d.db.Close()
	if err != nil {
		logger.WithField("error", err.Error()).Error("Failed to close database connection")
	} else {
		logger.Info("Database connection closed successfully")
	}
	return err
}
