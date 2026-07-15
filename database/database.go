package database

import (
	"database/sql"
	"fmt"
	"go-rest-api/utils"

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
	Driver = utils.GetEnv("DB_DRIVER", "sqlite")

	var err error
	switch Driver {
	case "postgres":
		DB, err = sql.Open("postgres", postgresDSN())
	case "sqlite":
		DB, err = sql.Open("sqlite3", utils.GetEnv("DB_PATH", "./api.db"))
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
		utils.GetEnv("DB_HOST", "localhost"),
		utils.GetEnv("DB_PORT", "5432"),
		utils.GetEnv("DB_USER", "postgres"),
		utils.GetEnv("DB_PASSWORD", "postgres"),
		utils.GetEnv("DB_NAME", "go_rest_api"),
		utils.GetEnv("DB_SSLMODE", "disable"),
	)
}

// CreateTables creates the schema needed by the app, using the syntax
// appropriate for the selected driver.
func CreateTables() {
	var createTableSQL string
	switch Driver {
	case "postgres":
		// users must be created before events: events.user_id has a FOREIGN
		// KEY on users(id), and Postgres validates that at CREATE TABLE time.
		createTableSQL = `
			CREATE TABLE IF NOT EXISTS users (
				id SERIAL PRIMARY KEY,
				email VARCHAR(100) NOT NULL UNIQUE,
				password VARCHAR(255) NOT NULL
		);
			CREATE TABLE IF NOT EXISTS events (
				id SERIAL PRIMARY KEY,
				name VARCHAR(100) NOT NULL,
				description TEXT,
				location VARCHAR(100) NOT NULL,
				date_time TIMESTAMP,
				user_id INTEGER,
				FOREIGN KEY (user_id) REFERENCES users(id) on DELETE CASCADE
		);
			CREATE TABLE IF NOT EXISTS registrations (
				id SERIAL PRIMARY KEY,
				event_id INTEGER NOT NULL,
				user_id INTEGER NOT NULL,
				FOREIGN KEY (event_id) REFERENCES events(id) on DELETE CASCADE,
				FOREIGN KEY (user_id) REFERENCES users(id) on DELETE CASCADE,
				UNIQUE (event_id, user_id)
		);
		`
	default:
		createTableSQL = `
			CREATE TABLE IF NOT EXISTS users (
				"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				"email" VARCHAR(100) NOT NULL UNIQUE,
				"password" VARCHAR(255) NOT NULL
		);
			CREATE TABLE IF NOT EXISTS events (
				"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				"name" VARCHAR(100) NOT NULL,
				"description" TEXT,
				"location" VARCHAR(100) NOT NULL,
				"date_time" DATETIME,
				"user_id" INTEGER,
				FOREIGN KEY (user_id) REFERENCES users(id) on DELETE CASCADE
		);
			CREATE TABLE IF NOT EXISTS registrations (
				"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				"event_id" INTEGER NOT NULL,
				"user_id" INTEGER NOT NULL,
				FOREIGN KEY (event_id) REFERENCES events(id) on DELETE CASCADE,
				FOREIGN KEY (user_id) REFERENCES users(id) on DELETE CASCADE,
				UNIQUE (event_id, user_id)
		);
		`
	}

	// DB.Prepare only allows a single statement (Postgres uses the extended
	// query protocol for it); DB.Exec uses the simple protocol, which allows
	// the multiple semicolon-separated CREATE TABLE statements above.
	if _, err := DB.Exec(createTableSQL); err != nil {
		panic(err)
	}
}
