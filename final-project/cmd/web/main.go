package main

import (
	"log"
	"os"
	"sync"

	"github.com/MatsuoTakuro/final-project/data"
)

// NOTE: of course, originally, this should be in an environment variable or config file.
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
	asyncJob := sync.WaitGroup{}

	// set up the server
	srv := Server{
		Session:   session,
		DB:        db,
		InfoLog:   infoLogger,
		ErrorLog:  errLogger,
		Models:    data.New(db),
		AsyncJob:  &asyncJob,
		AsyncErr:  make(chan error),
		StopAsync: make(chan bool),
	}

	// set up mail
	srv.initMailer()
	go srv.listenForMail()

	// listen for signals
	go srv.listenForShutdown()

	// listen for errors
	go srv.listenForAsyncJobErrors()

	// listen for web connections
	srv.serve()
}
