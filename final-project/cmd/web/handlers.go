package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"

	"github.com/MatsuoTakuro/final-project/data"
)

const (
	HOME_PAGE     = "home.page.gohtml"
	LOGIN_PAGE    = "login.page.gohtml"
	REGISTER_PAGE = "register.page.gohtml"
	PLANS_PAGE    = "plans.page.gohtml"
)

const (
	INVALID_CREDS_MSG            = "Invalid credentials."
	SUCCESSFUL_LOGIN_MSG         = "Successful login!"
	UNSUCCESSFUL_CREATE_USER_MSG = "Unable to create user."
	CONFIRMATION_EMAIL_SENT_MSG  = "Confirmation email sent. Check your email."
	INVALID_TOKEN_MSG            = "Invalid token."
	NOT_FOUND_USER_MSG           = "No user found."
	UNSUCCESSFUL_UPDATE_USER_MSG = "Unable to update user."
	ACCOUNT_ACTIVATED_MSG        = "Account activated. You can now log in."
	NEED_TO_LOGIN_FOR_PLANS_MSG  = "You must be logged in to view this page."
	ERROR_RENEW_TOKEN_MSG        = "error renewing token: %w"
	ERROR_PARSE_FORM_MSG         = "error parsing form: %w"
	ERROR_GET_ALL_PLANS_MSG      = "error getting all plans: %w"
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
		s.ErrorLog.Println(fmt.Errorf(ERROR_RENEW_TOKEN_MSG, err))
	}

	err = r.ParseForm()
	if err != nil {
		s.ErrorLog.Println(fmt.Errorf(ERROR_PARSE_FORM_MSG, err))
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
	http.Redirect(w, r, HOME_PATH, http.StatusSeeOther)
}

func (s *Server) Logout(w http.ResponseWriter, r *http.Request) {
	// clean up session
	_ = s.Session.Destroy(r.Context())
	_ = s.Session.RenewToken(r.Context()) // renew the session token every time the user logs out

	http.Redirect(w, r, HOME_PATH, http.StatusSeeOther)
}

func (s *Server) RegisterPage(w http.ResponseWriter, r *http.Request) {
	s.render(w, r, REGISTER_PAGE, nil)
}

func (s *Server) RegisterUser(w http.ResponseWriter, r *http.Request) {
	// create a user
	err := r.ParseForm()
	if err != nil {
		s.ErrorLog.Println(fmt.Errorf(ERROR_PARSE_FORM_MSG, err))
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
	activateURL := &url.URL{
		Scheme:   "http",
		Host:     r.Host,
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
	gotURL := &url.URL{
		Scheme:   "http",
		Host:     r.Host,
		Path:     r.URL.Path,
		RawQuery: r.URL.RawQuery,
	}

	if ok := VerifyToken(gotURL.String()); !ok {
		s.Session.Put(r.Context(), ERROR_CTX, INVALID_TOKEN_MSG)
		http.Redirect(w, r, HOME_PATH, http.StatusSeeOther)
		return
	}

	u, err := s.Models.User.GetByEmail(r.URL.Query().Get(EMAIL_ATTR))
	if err != nil {
		s.Session.Put(r.Context(), ERROR_CTX, NOT_FOUND_USER_MSG)
		http.Redirect(w, r, HOME_PATH, http.StatusSeeOther)
		return
	}

	u.IsActive = data.Active
	if err := u.Update(); err != nil {
		s.Session.Put(r.Context(), ERROR_CTX, UNSUCCESSFUL_UPDATE_USER_MSG)
		http.Redirect(w, r, HOME_PATH, http.StatusSeeOther)
		return
	}

	s.Session.Put(r.Context(), FLASH_CTX, ACCOUNT_ACTIVATED_MSG)
	http.Redirect(w, r, LOGIN_PATH, http.StatusSeeOther)
}

func (s *Server) SubcribeToPlan(w http.ResponseWriter, r *http.Request) {
	// get the id of the plan that is chosen

	// get the plan from the database

	// get the user from the session

	// generate an invoice

	// send an email with the invoice attached

	// generate a manual

	// send an email with the manual attached

	// subscribe the user to an account

	// redirect
}

func (s *Server) ListOfPlans(w http.ResponseWriter, r *http.Request) {
	if !s.Session.Exists(r.Context(), USER_ID_CTX) {
		s.Session.Put(r.Context(), WARNING_CTX, NEED_TO_LOGIN_FOR_PLANS_MSG)
		http.Redirect(w, r, LOGIN_PATH, http.StatusTemporaryRedirect)
		/* NOTE: status code 307 (vs 303)
			In the case of a 307 Temporary Redirect, the client should continue to use the original URL for future requests.
		This means that the client should use the same URL that was used in the original request, including the same HTTP method, headers, and body.

			In the example code you provided, the original URL for future requests is the URL that the client used to access the listOfPlans handler.
		When the client is redirected to the login page, the http.Redirect function is called with the original request (r) and a 307 Temporary Redirect status code.
		This ensures that the client continues to use the original URL for future requests, including any query parameters, headers, and request body.
		*/
		return
	}

	plans, err := s.Models.Plan.GetAll()
	if err != nil {
		s.ErrorLog.Println(fmt.Errorf(ERROR_GET_ALL_PLANS_MSG, err))
		return
	}

	dataMap := make(map[string]any)
	dataMap[PLANS_ATTR] = plans

	s.render(w, r, PLANS_PAGE, &TemplateData{
		Data: dataMap,
	})
}
