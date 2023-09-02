package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"

	"github.com/MatsuoTakuro/final-project/data"
	mail "github.com/xhit/go-simple-mail/v2"
)

// XXX_PAGE is the name of the template gohtml file to render for the page
const (
	HOME_PAGE     = "home.page.gohtml"
	LOGIN_PAGE    = "login.page.gohtml"
	REGISTER_PAGE = "register.page.gohtml"
	PLANS_PAGE    = "plans.page.gohtml"
)

// XXX_MSG is the message to display to the user or to log for you
const (
	INVALID_CREDS_MSG            = "Invalid credentials."
	SUCCESSFUL_LOGIN_MSG         = "Successful login!"
	UNSUCCESSFUL_CREATE_USER_MSG = "Unable to create user."
	CONFIRMATION_EMAIL_SENT_MSG  = "Confirmation email sent. Check your email."
	INVALID_TOKEN_MSG            = "Invalid token."
	NOT_FOUND_USER_BY_EMAIL_MSG  = "No user found with that email."
	NOT_FOUND_USER_BY_ID_MSG     = "No user found with that id."
	UNSUCCESSFUL_UPDATE_USER_MSG = "Unable to update user."
	ACCOUNT_ACTIVATED_MSG        = "Account activated. You can now log in."
	NEED_TO_LOGIN_FOR_PLANS_MSG  = "You must be logged in to view this page."
	ERROR_RENEW_TOKEN_MSG        = "error renewing token: %w"
	ERROR_PARSE_FORM_MSG         = "error parsing form: %w"
	ERROR_GET_ALL_PLANS_MSG      = "error getting all plans: %w"
	LOGIN_FIRST_MSG              = "Log in first!"
	UNSUCCESSFUL_FIND_PLAN_MSG   = "Unable to find plan."
	SUCCESSFUL_SUBSCRIBE_MSG     = "Subscribed!"
	UNSUCCESSFUL_SUBSCRIBE_MSG   = "Unable to subscribe."
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
		s.Session.Put(r.Context(), ERROR_CTX, NOT_FOUND_USER_BY_EMAIL_MSG)
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
	id := r.URL.Query().Get(PLAN_ID_CTX)
	planID, _ := strconv.Atoi(id)

	// get the plan from the database
	plan, err := s.Models.Plan.GetOne(planID)
	if err != nil {
		s.Session.Put(r.Context(), ERROR_CTX, UNSUCCESSFUL_FIND_PLAN_MSG)
		http.Redirect(w, r, membersPlanPath, http.StatusSeeOther)
		return
	}

	// get the user from the session
	user, ok := s.Session.Get(r.Context(), USER_CTX).(data.User)
	if !ok {
		s.Session.Put(r.Context(), ERROR_CTX, LOGIN_FIRST_MSG)
		http.Redirect(w, r, LOGIN_PATH, http.StatusSeeOther)
		return
	}

	// generate an invoice and email it
	s.AsyncJob.Add(1) // increment counter every time a new invoice is generated
	go func() {
		defer s.AsyncJob.Done() // decrement counter every time an invoice is generated and passed to the mailer to send

		invoice, err := s.getInvoice(user, plan)
		if err != nil {
			s.AsyncErr <- err
		}

		msg := Message{
			To:       user.Email,
			Subject:  "Your invoice",
			Data:     invoice,
			Template: "invoice",
		}
		s.sendEmail(msg)
	}()

	// generate a manual
	s.AsyncJob.Add(1) // increment counter every time a new manual is generated
	go func() {
		defer s.AsyncJob.Done() // decrement counter every time a manual is generated and passed to the mailer to send

		pdf := s.generateManual(user, plan)
		filePath := fmt.Sprintf(MANUAL_OUTPUT_TEMP_PATH, user.ID)
		err := pdf.OutputFileAndClose(filePath)
		if err != nil {
			s.AsyncErr <- err
			return
		}

		msg := Message{
			To:      user.Email,
			Subject: "Your manual",
			Data:    "Your user manual is attached",
			Attachments: []*mail.File{
				{
					Name:     MANUAL_ATTCH_NAME,
					FilePath: filePath,
				},
			},
		}

		s.sendEmail(msg)
	}()

	// subscribe the user to a plan
	err = s.Models.Plan.SubscribeUserToPlan(user, *plan)
	if err != nil {
		s.Session.Put(r.Context(), ERROR_CTX, UNSUCCESSFUL_SUBSCRIBE_MSG)
		http.Redirect(w, r, membersPlanPath, http.StatusSeeOther)
		return
	}

	u, err := s.Models.User.GetOne(user.ID)
	if err != nil {
		s.Session.Put(r.Context(), ERROR_CTX, NOT_FOUND_USER_BY_ID_MSG)
		http.Redirect(w, r, membersPlanPath, http.StatusSeeOther)
		return
	}

	s.Session.Put(r.Context(), USER_CTX, u)

	// redirect
	s.Session.Put(r.Context(), FLASH_CTX, SUCCESSFUL_SUBSCRIBE_MSG)
	http.Redirect(w, r, membersPlanPath, http.StatusSeeOther)
}

func (s *Server) ListOfPlans(w http.ResponseWriter, r *http.Request) {
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
