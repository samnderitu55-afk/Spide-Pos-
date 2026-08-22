package handlers

import (
    "html/template"
    "net/http"
    "path/filepath"
)

var (
    dashboardTemplate *template.Template
    posTemplate       *template.Template
    loginTemplate     *template.Template
    directorTemplate  *template.Template
)

func init() {
    dashboardTemplate = template.Must(template.ParseFiles(filepath.Join("internal", "templates", "dashboard.html")))
    posTemplate = template.Must(template.ParseFiles(filepath.Join("internal", "templates", "pos.html")))
    loginTemplate = template.Must(template.ParseFiles(filepath.Join("internal", "templates", "login.html")))
    directorTemplate = template.Must(template.ParseFiles(filepath.Join("internal", "templates", "director.html")))
}

func ServeDashboard(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.NotFound(w, r)
        return
    }
    w.Header().Set("Content-Type", "text/html")
    dashboardTemplate.Execute(w, nil)
}

func ServePOS(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    posTemplate.Execute(w, nil)
}

func ServeLogin(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    loginTemplate.Execute(w, nil)
}

func ServeDirector(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    directorTemplate.Execute(w, nil)
}




