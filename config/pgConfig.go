package config

import (
	"flag"
	"fmt"
)

type DatabaseConfig interface {
	GetConnectionString() string
}

// PostgresConfig represents the configuration needed to connect to a PostgreSQL database
type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"username"`
	Password string `yaml:"password"`
	DBName   string `yaml:"db_name"`
}

func (pc *PostgresConfig) GetConnectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		pc.Host, pc.Port, pc.User, pc.Password, pc.DBName)
}

// NewPostgresConfigFromFlags fetches the Postgres configuration using flags
func GetPostgresConfig() *PostgresConfig {
	return &PostgresConfig{
		Host:     *flag.String("POSTGRES_HOST", "localhost", "Postgres host"),
		Port:     *flag.String("POSTGRES_PORT", "5432", "Postgres port"),
		User:     *flag.String("POSTGRES_USER", "postgres", "Postgres user"),
		Password: *flag.String("POSTGRES_PASSWORD", "postgres", "Postgres password"),
		DBName:   *flag.String("POSTGRES_NAME", "postgres", "Postgres database name"),
	}
}
