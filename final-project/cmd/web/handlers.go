package main

import "net/http"

const (
	HOME_PAGE = "home.page.gohtml"
)

func (s *Server) HomePage(w http.ResponseWriter, r *http.Request) {
	s.render(w, r, HOME_PAGE, nil)
}
