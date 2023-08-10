package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/alexedwards/scs/v2"
)

type Server struct {
	Session  *scs.SessionManager
	DB       *sql.DB
	InfoLog  *log.Logger
	ErrorLog *log.Logger
	Wait     *sync.WaitGroup
}

func (s *Server) serve() {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", WEB_PORT),
		Handler: s.routes(),
	}

	s.InfoLog.Println("Starting web server...")
	if err := srv.ListenAndServe(); err != nil {
		log.Panic(err)
	}
}
