package repositories

import (
	"database/sql"
	"log"
)

func ConnectPostgres(dsn string) *sql.DB {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to connect to Postgres: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping Postgres: %v", err)
	}
	return db
}
