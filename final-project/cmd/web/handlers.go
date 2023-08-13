package main

import (
	"net/http"
)

const (
	HOME_PAGE     = "home.page.gohtml"
	LOGIN_PAGE    = "login.page.gohtml"
	REGISTER_PAGE = "register.page.gohtml"
)

const (
	INVALID_CREDS_MSG    = "Invalid credentials."
	SUCCESSFUL_LOGIN_MSG = "Successful login!"
)

func (s *Server) HomePage(w http.ResponseWriter, r *http.Request) {
	s.render(w, r, HOME_PAGE, nil)
}

func (s *Server) LoginPage(w http.ResponseWriter, r *http.Request) {
	s.render(w, r, LOGIN_PAGE, nil)
}

func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	err := s.Session.RenewToken(r.Context()) // renew the session token every time the user logs in
	if err != nil {
		s.ErrorLog.Println(err)
	}

	err = r.ParseForm()
	if err != nil {
		s.ErrorLog.Println(err)
	}

	email := r.Form.Get(EMAIL_ATTR)
	password := r.Form.Get(PASSWORD_ATTR)

	user, err := s.Models.User.GetByEmail(email)
	if err != nil {
		s.Session.Put(r.Context(), ERROR_CTX, INVALID_CREDS_MSG)
		http.Redirect(w, r, LOGIN_PATH, http.StatusSeeOther)
		return
	}

	isValidPassword, err := user.PasswordMatches(password)
	if err != nil {
		s.Session.Put(r.Context(), ERROR_CTX, INVALID_CREDS_MSG)
		http.Redirect(w, r, LOGIN_PATH, http.StatusSeeOther)
		return
	}

	if !isValidPassword {
		s.Session.Put(r.Context(), ERROR_CTX, INVALID_CREDS_MSG)
		http.Redirect(w, r, LOGIN_PATH, http.StatusSeeOther)
		return
	}

	// log user in
	s.Session.Put(r.Context(), USER_ID_CTX, user.ID)
	s.Session.Put(r.Context(), USER_CTX, user)
	s.Session.Put(r.Context(), FLASH_CTX, SUCCESSFUL_LOGIN_MSG)

	// redirect the user to the home page
	http.Redirect(w, r, HOGE_PATH, http.StatusSeeOther)
}

func (s *Server) Logout(w http.ResponseWriter, r *http.Request) {
	// clean up session
	_ = s.Session.Destroy(r.Context())
	_ = s.Session.RenewToken(r.Context()) // renew the session token every time the user logs out
}

func (s *Server) RegisterPage(w http.ResponseWriter, r *http.Request) {
	s.render(w, r, REGISTER_PAGE, nil)
}

func (s *Server) RegisterAccount(w http.ResponseWriter, r *http.Request) {
	// create a user

	// send an activation email

	// subscbribe the user to an account
}

func (s *Server) ActivateAccount(w http.ResponseWriter, r *http.Request) {
	// validate url

	// generate an invoice

	// send an email with attachments

	// send an email with the invoice attached
}
