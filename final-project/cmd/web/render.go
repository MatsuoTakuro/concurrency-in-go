package main

import (
	"net/http"
	"path/filepath"
	"text/template"
	"time"
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
	// User *data.User
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
		s.ErrorLog.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// write the applied output to the response writer
	if err := tmpl.Execute(w, s.UpdateDefaultData(td, r)); err != nil {
		s.ErrorLog.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

//
func (s *Server) UpdateDefaultData(td *TemplateData, r *http.Request) *TemplateData {
	td.Flash = s.Session.PopString(r.Context(), FLASH_KEY)
	td.Warning = s.Session.PopString(r.Context(), WARNING_KEY)
	td.Error = s.Session.PopString(r.Context(), ERROR_KEY)
	if s.IsAuthenticated(r) {
		td.Authenticated = true
		// TODO - get more user information
	}
	td.Now = time.Now()

	return td
}

func (s *Server) IsAuthenticated(r *http.Request) bool {
	return s.Session.Exists(r.Context(), USER_ID_KEY)
}
