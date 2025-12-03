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

// ServeIndex serves the main page
func (h *WebHandler) ServeIndex(w http.ResponseWriter, _ *http.Request) error {
	return h.templates.ExecuteTemplate(w, "index.html", nil)
}

// ServeURLTab serves the URL shortener tab
func (h *WebHandler) ServeURLTab(w http.ResponseWriter, _ *http.Request) error {
	return h.templates.ExecuteTemplate(w, "url-shortener.html", nil)
}

// ServeImageTab serves the image sharing tab
func (h *WebHandler) ServeImageTab(w http.ResponseWriter, _ *http.Request) error {
	return h.templates.ExecuteTemplate(w, "image-sharing.html", nil)
}

// ServeSecretTab serves the secret sharing tab
func (h *WebHandler) ServeSecretTab(w http.ResponseWriter, _ *http.Request) error {
	return h.templates.ExecuteTemplate(w, "secret-sharing.html", nil)
}

// ServeCodeTab serves the code sharing tab
func (h *WebHandler) ServeCodeTab(w http.ResponseWriter, _ *http.Request) error {
	return h.templates.ExecuteTemplate(w, "code-sharing.html", nil)
}

// ServeIPTab serves the IP detection tab
func (h *WebHandler) ServeIPTab(w http.ResponseWriter, _ *http.Request) error {
	return h.templates.ExecuteTemplate(w, "ip-detection.html", nil)
}

// DetectIP returns the user's IP address as HTML
func (h *WebHandler) DetectIP(w http.ResponseWriter, _ *http.Request) error {
	// Call the internal API
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

	// Render IP result template
	return h.templates.ExecuteTemplate(w, "ip-result.html", map[string]string{
		"IPAddress": ipResp.IP,
	})
}
