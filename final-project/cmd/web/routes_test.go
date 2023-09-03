package main

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
)

var routes = []string{
	HOME_PATH,
	LOGIN_PATH,
	LOGOUT_PATH,
	REGISTER_PATH,
	ACTIVATE_PATH,
	MembersPlanPath,
	MembersSubscribePath,
}

var _ http.Handler = (chi.Router)(nil)

func Test_Routes_Exist(t *testing.T) {
	testRouter := testServer.routes()
	chiRouter := testRouter.(chi.Router)
	for _, r := range routes {
		routeExists(t, chiRouter, r)
	}
}

func routeExists(t *testing.T, router chi.Router, route string) {
	var found bool

	_ = chi.Walk(router, func(method, foundRoute string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		if route == foundRoute {
			found = true
		}
		return nil
	})

	if !found {
		t.Errorf("did not find %s in registered routes", route)
	}
}
