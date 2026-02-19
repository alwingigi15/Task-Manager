package config

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

var (
	dbInstance   *sql.DB    // Singleton instance of the database connection
	once         sync.Once  // Ensures InitDB is executed only once
	connectionMu sync.Mutex // Protects reinitialization
)

func Dbconnection() (*sql.DB, error) {
	connectionMu.Lock()
	defer connectionMu.Unlock()

	if dbInstance == nil {
		log.Println("Database connection is not initialized. Initializing now.")
		if err := InitDB(ReadDbConnection()); err != nil {
			return nil, fmt.Errorf("failed to initialize DB: %w", err)
		}
	}

	const maxRetries = 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := dbInstance.Ping()
		if err == nil {
			return dbInstance, nil
		}
		log.Printf("Database ping failed (attempt %d/%d): %v", attempt, maxRetries, err)
		if attempt < maxRetries {
			// Reinitialize only if necessary
			if err := InitDB(ReadDbConnection()); err != nil {
				log.Printf("Failed to reinitialize DB (attempt %d/%d): %v", attempt, maxRetries, err)
			}
			time.Sleep(time.Duration(attempt*100) * time.Millisecond)
		}
	}
	return nil, fmt.Errorf("failed to ping database after %d attempts", maxRetries)
}

func InitDB(connectionString string) error {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	const maxRetries = 3
	var db *sql.DB
	//var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		once.Do(func() {
			log.Printf("Initializing database connection pool with host=%s, dbname=%s, port=%s",
				cfg.DbHost, cfg.DbName, cfg.DbPort)

			db, err = sql.Open("postgres", connectionString)
			if err != nil {
				log.Printf("Failed to open database connection (attempt %d/%d): %v", attempt, maxRetries, err)
				return
			}

			// Configure connection pool settings
			db.SetMaxOpenConns(18) // Increased for production load
			db.SetMaxIdleConns(8) // Adjusted for balance
			db.SetConnMaxLifetime(2 * time.Minute) // Reduced to avoid stale connections
			db.SetConnMaxIdleTime(15 * time.Second)

			// Test the connection
			if err = db.Ping(); err != nil {
				db.Close()
				log.Printf("Failed to ping database (attempt %d/%d): %v", attempt, maxRetries, err)
				return
			}

			dbInstance = db
			log.Println("Database connection pool initialized successfully.")
		})
		if err == nil && dbInstance != nil {
			return nil
		}
		if attempt < maxRetries {
			time.Sleep(time.Duration(attempt*100) * time.Millisecond)
		}
	}
	return fmt.Errorf("failed to initialize database after %d attempts: %w", maxRetries, err)
}

func CloseDB() {
	connectionMu.Lock()
	defer connectionMu.Unlock()
	if dbInstance != nil {
		if err := dbInstance.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		} else {
			log.Println("Database connection pool closed successfully.")
		}
		dbInstance = nil
		once = sync.Once{} // Reset for potential reinitialization
	}
}

func ReadDbConnection() string {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	psqlInfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DbHost, cfg.DbUser, cfg.DbPassword, cfg.DbName, cfg.DbPort, cfg.SSLmode)

	return psqlInfo
}
