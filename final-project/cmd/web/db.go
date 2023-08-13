package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func initDB() *sql.DB {
	conn := connectToDB()
	if conn == nil {
		log.Panic("can't connect to database")
	}
	return conn
}

func connectToDB() *sql.DB {
	var counts uint8

	dsn := os.Getenv("DSN")

	for {
		conn, err := openDB(dsn)
		if err != nil {
			log.Println("postgres not yet ready...")
		} else {
			log.Print("connected to database!")
			return conn
		}

		if counts > 10 {
			return nil
		}

		log.Println("Backing off for 1 second")
		time.Sleep(1 * time.Second)
		counts++

		continue
	}
}

func openDB(dsn string) (*sql.DB, error) {
	// NOTE: https://pkg.go.dev/database/sql#Open
	// The returned DB is safe for concurrent use by multiple goroutines and maintains its own pool of idle connections.
	// Thus, the Open function should be called just once. It is rarely necessary to close a DB.
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
