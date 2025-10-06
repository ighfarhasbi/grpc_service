package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

func NewOrderPostgres(dbUrl string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		return nil, fmt.Errorf("error opening order database: %w", err)
	}

	// Set konfigurasi pool connection
	db.SetMaxOpenConns(25)                  // maksimal 25 koneksi aktif
	db.SetMaxIdleConns(10)                  // simpan 10 koneksi idle
	db.SetConnMaxLifetime(5 * time.Minute)  // koneksi akan diganti tiap 5 menit
	db.SetConnMaxIdleTime(30 * time.Second) // koneksi idle akan ditutup setelah 30 detik

	// Test koneksi
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error ping order database: %w", err)
	}

	log.Println("Connected to Order PostgreSQL")
	return db, nil
}
