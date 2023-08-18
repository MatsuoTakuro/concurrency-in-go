package main

import (
	"log"
	"os"
	"sync"

	"github.com/MatsuoTakuro/final-project/data"
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
	sentMail := sync.WaitGroup{}

	// set up the server
	srv := Server{
		Session:  session,
		DB:       db,
		InfoLog:  infoLogger,
		ErrorLog: errLogger,
		SentMail: &sentMail,
		Models:   data.New(db),
	}

	// set up mail
	srv.initMailer()
	go srv.listenForMail()

	// listen for signals
	go srv.listenForShutdown()

	// listen for web connections
	srv.serve()
}
