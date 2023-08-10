package main

import (
	"log"
	"os"
	"sync"
)

const WEB_PORT = "80"

func main() {
	// connect to the database
	db := initDB()

	// create sessions
	session := initSession()

	// create loggers
	infoLogger := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errLogger := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// create channels

	// create waitgroup
	wg := sync.WaitGroup{}

	// set up the server
	srv := Server{
		Session:  session,
		DB:       db,
		InfoLog:  infoLogger,
		ErrorLog: errLogger,
		Wait:     &wg,
	}

	// set up mail

	// listen for web connections
	srv.serve()
}
