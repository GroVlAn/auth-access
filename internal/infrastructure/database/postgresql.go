package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type PostgresSettings struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

func NewPostgresqlDB(settings PostgresSettings) (*sqlx.DB, error) {
	db, err := sqlx.Open(
		"postgres",
		fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			settings.Host, settings.Port, settings.Username, settings.Password, settings.DBName, settings.SSLMode))
	if err != nil {
		return nil, fmt.Errorf("connection to postgres db: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging postgres db: %w", err)
	}

	return db, nil
}
