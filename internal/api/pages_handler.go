package api

import (
	"html/template"
	"net/http"
)

var pageTemplates = template.Must(template.ParseGlob("web/templates/*.html"))

func DashboardPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := pageTemplates.ExecuteTemplate(w, "layout.html", nil); err != nil {
		http.Error(w, "erreur de rendu de la page", http.StatusInternalServerError)
	}
}
