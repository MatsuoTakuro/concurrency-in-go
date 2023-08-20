package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"text/template"
	"time"

	"github.com/MatsuoTakuro/final-project/data"
)

const (
	TEMPLATE_PATH = "./cmd/web/templates"
)

const (
	BASE_LAYOUT    = "base.layout.gohtml"
	PARTIAL_HEADER = "header.partial.gohtml"
	PARTIAL_NAVBAR = "navbar.partial.gohtml"
	PARTIAL_FOOTER = "footer.partial.gohtml"
	PARTIAL_ALERTS = "alerts.partial.gohtml"
)

const (
	UNSUCCESSFUL_GET_SESSION_USER_MSG = "can't get user from session"
	ERROR_PARSE_TEMPLATE_FILES_MSG    = "error parsing template files: %w"
	ERROR_EXECUTING_TEMPLATE_MSG      = "error executing template: %w"
)

var Partials = []string{
	filepath.Join(TEMPLATE_PATH, BASE_LAYOUT),
	filepath.Join(TEMPLATE_PATH, PARTIAL_HEADER),
	filepath.Join(TEMPLATE_PATH, PARTIAL_NAVBAR),
	filepath.Join(TEMPLATE_PATH, PARTIAL_FOOTER),
	filepath.Join(TEMPLATE_PATH, PARTIAL_ALERTS),
}

// TemplateData holds data sent from handlers to templates
type TemplateData struct {
	StringMap     map[string]string
	IntMap        map[string]int
	FloatMap      map[string]float64
	Data          map[string]any
	Flash         string
	Warning       string
	Error         string
	Authenticated bool
	Now           time.Time
	User          *data.User
}

func (s *Server) render(
	w http.ResponseWriter,
	r *http.Request,
	targetHTML string,
	td *TemplateData,
) {

	var baseTmpls []string
	baseTmpls = append(baseTmpls, filepath.Join(TEMPLATE_PATH, targetHTML))
	baseTmpls = append(baseTmpls, Partials...)

	if td == nil {
		td = &TemplateData{}
	}

	// parse the template files
	tmpl, err := template.ParseFiles(baseTmpls...)
	if err != nil {
		s.ErrorLog.Println(fmt.Errorf(ERROR_PARSE_TEMPLATE_FILES_MSG, err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// write the applied output to the response writer
	if err := tmpl.Execute(w, s.UpdateDefaultData(td, r)); err != nil {
		s.ErrorLog.Println(fmt.Errorf(ERROR_EXECUTING_TEMPLATE_MSG, err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

//
func (s *Server) UpdateDefaultData(td *TemplateData, r *http.Request) *TemplateData {
	td.Flash = s.Session.PopString(r.Context(), FLASH_CTX)
	td.Warning = s.Session.PopString(r.Context(), WARNING_CTX)
	td.Error = s.Session.PopString(r.Context(), ERROR_CTX)
	if s.IsAuthenticated(r) {
		td.Authenticated = true
		u, ok := s.Session.Get(r.Context(), USER_CTX).(data.User)
		if !ok {
			s.ErrorLog.Println(UNSUCCESSFUL_GET_SESSION_USER_MSG)
		} else {
			td.User = &u
		}
	}
	td.Now = time.Now()

	return td
}

func (s *Server) IsAuthenticated(r *http.Request) bool {
	return s.Session.Exists(r.Context(), USER_ID_CTX)
}
