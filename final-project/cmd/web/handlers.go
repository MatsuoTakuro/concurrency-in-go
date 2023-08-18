package main

import (
	"html/template"
	"net/http"
	"net/url"

	"github.com/MatsuoTakuro/final-project/data"
)

const (
	HOME_PAGE     = "home.page.gohtml"
	LOGIN_PAGE    = "login.page.gohtml"
	REGISTER_PAGE = "register.page.gohtml"
)

const (
	INVALID_CREDS_MSG            = "Invalid credentials."
	SUCCESSFUL_LOGIN_MSG         = "Successful login!"
	UNSUCCESSFUL_CREATE_USER_MSG = "Unable to create user."
	CONFIRMATION_EMAIL_SENT_MSG  = "Confirmation email sent. Check your email."
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
		msg := Message{
			To:      email,
			Subject: "Failed log in attempt",
			Data:    "Invalid login attempt!",
		}
		s.sendEmail(msg)

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

func (s *Server) RegisterUser(w http.ResponseWriter, r *http.Request) {
	// create a user
	err := r.ParseForm()
	if err != nil {
		s.ErrorLog.Println(err)
	}

	// validate data
	// NOTE: - originally, here should validate the data by checking if the user already exists or not and other stuff

	// send an activation email
	u := data.User{
		Email:     r.Form.Get(EMAIL_ATTR),
		FirstName: r.Form.Get(FIRST_NAME_ATTR),
		LastName:  r.Form.Get(LAST_NAME_ATTR),
		Password:  r.Form.Get(PASSWORD_ATTR),
		IsActive:  data.Inactive,
		IsAdmin:   data.NotAdmin,
	}

	_, err = u.Insert(u)
	if err != nil {
		s.Session.Put(r.Context(), ERROR_CTX, UNSUCCESSFUL_CREATE_USER_MSG)
		http.Redirect(w, r, REGISTER_PATH, http.StatusSeeOther)
		return
	}

	// send an activation email
	q := url.Values{}
	q.Set(EMAIL_ATTR, u.Email)
	// NOTE: originally, scheme and host should be in an environment variable or config file.
	activateURL := &url.URL{
		Scheme:   "http",
		Host:     "localhost",
		Path:     ACTIVATE_PATH,
		RawQuery: q.Encode(),
	}
	signedURL := GenerateTokenFromString(activateURL.String())
	s.InfoLog.Println(signedURL)

	msg := Message{
		To:       u.Email,
		Subject:  "Activate your account",
		Template: CONFIRM_EMAIL,
		Data:     template.HTML(signedURL),
	}
	s.sendEmail(msg)

	s.Session.Put(r.Context(), FLASH_CTX, CONFIRMATION_EMAIL_SENT_MSG)
	http.Redirect(w, r, LOGIN_PATH, http.StatusSeeOther)
}

func (s *Server) ActivateUserAccount(w http.ResponseWriter, r *http.Request) {
	// validate url

	// generate an invoice

	// send an email with attachments

	// send an email with the invoice attached
}
