package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// DB is the shared database handle used by repository implementations.
var DB *sql.DB

// Driver is the driver selected via DB_DRIVER ("sqlite" or "postgres").
var Driver string

// InitDB opens the database connection selected by the DB_DRIVER env var.
// Defaults to sqlite so local/dev usage keeps working without extra setup.
func InitDB() {
	Driver = getEnv("DB_DRIVER", "sqlite")

	var err error
	switch Driver {
	case "postgres":
		DB, err = sql.Open("postgres", postgresDSN())
	case "sqlite":
		DB, err = sql.Open("sqlite3", getEnv("DB_PATH", "./api.db"))
	default:
		panic(fmt.Sprintf("unknown DB_DRIVER %q (want \"sqlite\" or \"postgres\")", Driver))
	}
	if err != nil {
		panic(err)
	}

	if err := DB.Ping(); err != nil {
		panic(err)
	}

	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(5)
}

func postgresDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "postgres"),
		getEnv("DB_NAME", "go_rest_api"),
		getEnv("DB_SSLMODE", "disable"),
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// CreateTables creates the schema needed by the app, using the syntax
// appropriate for the selected driver.
func CreateTables() {
	var createTableSQL string
	switch Driver {
	case "postgres":
		createTableSQL = `
			CREATE TABLE IF NOT EXISTS events (
				id SERIAL PRIMARY KEY,
				name VARCHAR(100) NOT NULL,
				description TEXT,
				location VARCHAR(100) NOT NULL,
				date_time TIMESTAMP,
				user_id INTEGER
		);`
	default:
		createTableSQL = `
			CREATE TABLE IF NOT EXISTS events (
				"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				"name" VARCHAR(100) NOT NULL,
				"description" TEXT,
				"location" VARCHAR(100) NOT NULL,
				"date_time" DATETIME,
				"user_id" INTEGER
		);`
	}

	statement, err := DB.Prepare(createTableSQL)
	if err != nil {
		panic(err)
	}
	statement.Exec()
}
