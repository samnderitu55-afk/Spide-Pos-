package handlers

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

// loadTemplate parses a template from disk on every call.
// This means changes to .html files show up on the next request —
// no server restart needed. The cost is ~1ms per page load, which is
// negligible for a POS with a handful of page views per minute.
func loadTemplate(name string) (*template.Template, error) {
	path := filepath.Join("internal", "templates", name)
	return template.ParseFiles(path)
}

// setNoCacheHeaders tells browsers and proxies to never cache HTML pages.
// Prevents the "I edited the template but the browser shows the old one" class of bug.
func setNoCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

func ServeDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	tmpl, err := loadTemplate("dashboard.html")
	if err != nil {
		log.Printf("❌ Failed to load dashboard template: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	setNoCacheHeaders(w)
	tmpl.Execute(w, nil)
}

func ServePOS(w http.ResponseWriter, r *http.Request) {
	tmpl, err := loadTemplate("pos.html")
	if err != nil {
		log.Printf("❌ Failed to load pos template: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	setNoCacheHeaders(w)
	tmpl.Execute(w, nil)
}

func ServeLogin(w http.ResponseWriter, r *http.Request) {
	tmpl, err := loadTemplate("login.html")
	if err != nil {
		log.Printf("❌ Failed to load login template: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	setNoCacheHeaders(w)
	tmpl.Execute(w, nil)
}

func ServeDirector(w http.ResponseWriter, r *http.Request) {
	tmpl, err := loadTemplate("director.html")
	if err != nil {
		log.Printf("❌ Failed to load director template: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	setNoCacheHeaders(w)
	tmpl.Execute(w, nil)
}
