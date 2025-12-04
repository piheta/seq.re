package web

import (
	"html/template"
	"net/http"

	"github.com/piheta/seq.re/config"
	"github.com/piheta/seq.re/internal/shared"
)

type WebHandler struct {
	templates *template.Template
	version   string
}

func NewWebHandler(version string) *WebHandler {
	tmpl := template.Must(template.ParseGlob("web/templates/partials/*.html"))
	tmpl = template.Must(tmpl.ParseFiles("web/templates/index.html"))

	return &WebHandler{
		templates: tmpl,
		version:   version,
	}
}

func (h *WebHandler) ServeIndex(w http.ResponseWriter, _ *http.Request) error {
	data := map[string]string{
		"ContactEmail": config.Config.ContactEmail,
		"Version":      h.version,
	}
	return h.templates.ExecuteTemplate(w, "index.html", data)
}

func (h *WebHandler) ServeURLTab(w http.ResponseWriter, _ *http.Request) error {
	return h.templates.ExecuteTemplate(w, "url-shortener.html", nil)
}

func (h *WebHandler) ServeImageTab(w http.ResponseWriter, _ *http.Request) error {
	return h.templates.ExecuteTemplate(w, "image-sharing.html", nil)
}

func (h *WebHandler) ServeSecretTab(w http.ResponseWriter, _ *http.Request) error {
	return h.templates.ExecuteTemplate(w, "secret-sharing.html", nil)
}

func (h *WebHandler) ServeCodeTab(w http.ResponseWriter, _ *http.Request) error {
	return h.templates.ExecuteTemplate(w, "code-sharing.html", nil)
}

func (h *WebHandler) ServeIPTab(w http.ResponseWriter, _ *http.Request) error {
	return h.templates.ExecuteTemplate(w, "ip-detection.html", nil)
}

func (h *WebHandler) DetectIP(w http.ResponseWriter, r *http.Request) error {
	ip := shared.GetIP(r)

	return h.templates.ExecuteTemplate(w, "ip-result.html", map[string]string{
		"IPAddress": ip,
	})
}
