package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"gomarketplace_api/config"
	"log"
	"time"
)

const maxRetries = 10
const dbMaxOpenConns = 20
const retryDelay = 5 * time.Second

func ConnectToPostgreSQL() (*sql.DB, error) {
	dbConfig := config.GetConfig()

	conStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.DBName)

	var db *sql.DB
	var err error
	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", conStr)
		if err != nil {
			log.Printf("Failed to connect to Postgres (attempt %d/%d): %v, %s", i+1, maxRetries, err, conStr)
			time.Sleep(retryDelay)
			continue
		}

		db.SetMaxOpenConns(dbMaxOpenConns)

		if err := db.Ping(); err != nil {
			log.Printf("Failed to ping Postgres db (attempt %d/%d): %v, %s", i+1, maxRetries, err, conStr)
			db.Close()
			time.Sleep(retryDelay)
			continue
		}

		log.Printf("Successfully connected to Postgres: %s", conStr)
		return db, nil
	}
	return nil, err
}
