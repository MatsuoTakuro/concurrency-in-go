package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	HOGE_PATH     = "/"
	LOGIN_PATH    = "/login"
	LOGOUT_PATH   = "/logout"
	REGISTER_PATH = "/register"
	ACTIVATE_PATH = "/activate"
)

func (s *Server) routes() http.Handler {

	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer) // recover from panic, log the panic error and return a 500 response
	mux.Use(s.SessionLoad)

	mux.Get(HOGE_PATH, s.HomePage)
	mux.Get(LOGIN_PATH, s.LoginPage)
	mux.Post(LOGIN_PATH, s.Login)
	mux.Get(LOGOUT_PATH, s.Logout)
	mux.Get(REGISTER_PATH, s.RegisterPage)
	mux.Post(REGISTER_PATH, s.RegisterUser)
	mux.Get(ACTIVATE_PATH, s.ActivateUserAccount)

	return mux
}
