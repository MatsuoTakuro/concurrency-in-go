package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) routes() http.Handler {

	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer) // recover from panic, log the panic error and return a 500 response
	mux.Use(s.SessionLoad)

	mux.Get("/", s.HomePage)

	return mux
}
