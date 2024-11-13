package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"gomarketplace_api/config"
	"log"
	"sync"
	"time"
)

const maxRetries = 10
const dbMaxOpenConns = 20
const retryDelay = 5 * time.Second

type PgConnector struct {
	config.DbConfig
	db *sql.DB
	mu sync.Mutex // Для защиты доступа к db
}

func NewPgConnector(dbConfig config.DbConfig) *PgConnector {
	return &PgConnector{DbConfig: dbConfig}
}
func (pg *PgConnector) Connect() (*sql.DB, error) {
	pg.mu.Lock()
	defer pg.mu.Unlock()

	if pg.db != nil {
		return pg.db, nil
	}

	var err error
	conStr := pg.GetConnectionString()

	for i := 0; i < maxRetries; i++ {
		pg.db, err = sql.Open("postgres", conStr)
		if err != nil {
			log.Printf("Failed to connect to Postgres (attempt %d/%d): %v, %s", i+1, maxRetries, err, conStr)
			time.Sleep(retryDelay)
			continue
		}

		pg.db.SetMaxOpenConns(dbMaxOpenConns)

		if err := pg.db.Ping(); err != nil {
			log.Printf("Failed to ping Postgres db (attempt %d/%d): %v, %s", i+1, maxRetries, err, conStr)
			pg.db.Close()
			time.Sleep(retryDelay)
			continue
		}

		log.Printf("Successfully connected to Postgres: %s", conStr)
		return pg.db, nil
	}
	return nil, err
}

func (pg *PgConnector) Ping() error {
	pg.mu.Lock()
	defer pg.mu.Unlock()

	if pg.db == nil {
		return fmt.Errorf("database connection is not established")
	}

	if err := pg.db.Ping(); err != nil {
		pg.db.Close()
		pg.db = nil
		return fmt.Errorf("ping failed: %w", err)
	}
	return nil
}
