package web

import (
	"net/http"

	"github.com/piheta/seq.re/config"
	"github.com/piheta/seq.re/internal/shared"
)

type WebHandler struct {
	templateService *shared.TemplateService
	version         string
}

func NewWebHandler(templateService *shared.TemplateService, version string) *WebHandler {
	return &WebHandler{
		templateService: templateService,
		version:         version,
	}
}

func (h *WebHandler) ServeIndex(w http.ResponseWriter, _ *http.Request) error {
	data := map[string]string{
		"ContactEmail": config.Config.ContactEmail,
		"Version":      h.version,
	}
	return h.templateService.RenderIndexTemplate(w, "index.html", data)
}

func (h *WebHandler) ServeURLTab(w http.ResponseWriter, _ *http.Request) error {
	return h.templateService.RenderIndexTemplate(w, "url-shortener.html", nil)
}

func (h *WebHandler) ServeImageTab(w http.ResponseWriter, _ *http.Request) error {
	return h.templateService.RenderIndexTemplate(w, "image-sharing.html", nil)
}

func (h *WebHandler) ServeSecretTab(w http.ResponseWriter, _ *http.Request) error {
	return h.templateService.RenderIndexTemplate(w, "secret-sharing.html", nil)
}

func (h *WebHandler) ServeCodeTab(w http.ResponseWriter, _ *http.Request) error {
	return h.templateService.RenderIndexTemplate(w, "code-sharing.html", nil)
}

func (h *WebHandler) ServeIPTab(w http.ResponseWriter, _ *http.Request) error {
	return h.templateService.RenderIndexTemplate(w, "ip-detection.html", nil)
}

func (h *WebHandler) DetectIP(w http.ResponseWriter, r *http.Request) error {
	ip := shared.GetIP(r)

	return h.templateService.RenderIndexTemplate(w, "ip-result.html", map[string]string{
		"IPAddress": ip,
	})
}
