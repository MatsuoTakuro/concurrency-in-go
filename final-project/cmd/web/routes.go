package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	HOME_PATH      = "/"
	LOGIN_PATH     = "/login"
	LOGOUT_PATH    = "/logout"
	REGISTER_PATH  = "/register"
	ACTIVATE_PATH  = "/activate"
	MEMBERS_PATH   = "/members"
	PLANS_PATH     = "/plans"
	SUBSCRIBE_PATH = "/subscribe"
)

var MembersPlanPath string = MEMBERS_PATH + PLANS_PATH
var MembersSubscribePath string = MEMBERS_PATH + SUBSCRIBE_PATH

func (s *Server) routes() http.Handler {

	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer) // recover from panic, log the panic error and return a 500 response
	mux.Use(s.SessionLoad)

	mux.Get(HOME_PATH, s.HomePage)
	mux.Get(LOGIN_PATH, s.LoginPage)
	mux.Post(LOGIN_PATH, s.Login)
	mux.Get(LOGOUT_PATH, s.Logout)
	mux.Get(REGISTER_PATH, s.RegisterPage)
	mux.Post(REGISTER_PATH, s.RegisterUser)
	mux.Get(ACTIVATE_PATH, s.ActivateUserAccount)

	// attach membershipRouter as a subrouter to root router
	mux.Mount(MEMBERS_PATH, s.membershipRouter())

	return mux
}

func (s *Server) membershipRouter() http.Handler {
	mux := chi.NewRouter()
	mux.Use(s.Auth)

	mux.Get(PLANS_PATH, s.ListOfPlans)
	mux.Get(SUBSCRIBE_PATH, s.SubcribeToPlan)

	return mux
}
