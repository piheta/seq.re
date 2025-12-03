package web

import (
	"encoding/json"
	"html/template"
	"io"
	"net/http"
)

type WebHandler struct {
	templates *template.Template
	apiClient *http.Client
}

func NewWebHandler() *WebHandler {
	tmpl := template.Must(template.ParseGlob("web/templates/partials/*.html"))
	tmpl = template.Must(tmpl.ParseFiles("web/templates/index.html"))

	return &WebHandler{
		templates: tmpl,
		apiClient: http.DefaultClient,
	}
}

func (h *WebHandler) ServeIndex(w http.ResponseWriter, _ *http.Request) error {
	return h.templates.ExecuteTemplate(w, "index.html", nil)
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

func (h *WebHandler) DetectIP(w http.ResponseWriter, _ *http.Request) error {
	resp, err := h.apiClient.Get("http://localhost:8080/api/ip")
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var ipResp struct {
		IP string `json:"ip"`
	}
	if err := json.Unmarshal(body, &ipResp); err != nil {
		return err
	}

	return h.templates.ExecuteTemplate(w, "ip-result.html", map[string]string{
		"IPAddress": ipResp.IP,
	})
}
