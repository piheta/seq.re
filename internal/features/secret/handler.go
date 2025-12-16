package secret

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/piheta/apicore/apierr"
	"github.com/piheta/apicore/response"
	"github.com/piheta/seq.re/config"
	s "github.com/piheta/seq.re/internal/shared"
)

type SecretHandler struct {
	secretService   *SecretService
	templateService *s.TemplateService
}

func NewSecretHandler(secretService *SecretService, templateService *s.TemplateService) *SecretHandler {
	return &SecretHandler{
		secretService:   secretService,
		templateService: templateService,
	}
}

// CreateSecret creates a new shortened URL.
// @Summary Create a shortened URL
// @Description Creates a new shortened URL from the provided original URL
// @Tags secret
// @Accept json
// @Produce json
// @Param request body SecretRequest true "Secret request with URL to shorten"
// @Success 201 {string} string "Shortened URL"
// @Failure 400 "Invalid request or URL format"
// @Failure 500 "Internal server error"
// @Router /api/secrets [post]
func (h *SecretHandler) CreateSecret(w http.ResponseWriter, r *http.Request) error {
	var secretReq SecretRequest
	if err := json.NewDecoder(r.Body).Decode(&secretReq); err != nil {
		return err
	}

	if err := s.Validate.Struct(&secretReq); err != nil {
		return err
	}

	secret, err := h.secretService.CreateSecret(secretReq.Data)
	if err != nil {
		return err
	}

	secretURL := fmt.Sprintf("%s%s/s/%s", config.Config.RedirectHost, config.Config.RedirectPort, secret.Short)

	// Check if this is an HTMX request (wants HTML response)
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := map[string]string{
			"URL":      secretURL,
			"ButtonID": "secret",
			"Warning":  "This link will self-destruct after being viewed once.",
		}
		return h.templateService.RenderResult(w, data)
	}

	return response.JSON(w, 201, secretURL)
}

// GetSecretByShort shows the one-time view page for the secret.
// @Summary Get secret information
// @Description Shows the one-time view page where users can reveal the secret
// @Tags secret
// @Param short path string true "Short code (6 characters)"
// @Success 200 {string} string "One-time view page"
// @Failure 404
// @Failure 422
// @Router /s/{short} [get]
func (h *SecretHandler) GetSecretByShort(w http.ResponseWriter, r *http.Request) error {
	short := r.PathValue("short")

	if len(short) != 6 {
		return s.MapError(w, r, apierr.NewError(422, "validation", "Invalid secret code"), h.templateService)
	}

	// Check if secret exists without consuming it
	exists, err := h.secretService.CheckSecretExists(short)
	if err != nil || !exists {
		return s.MapError(w, r, apierr.NewError(404, "not_found", "Secret not found or already viewed"), h.templateService)
	}

	if s.IsBrowser(r) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := map[string]string{
			"ID":   short,
			"Type": "secret",
		}
		return h.templateService.RenderOnetime(w, data)
	}

	// For API clients, immediately reveal the secret (backward compatibility)
	secret, err := h.secretService.GetSecret(short)
	if err != nil {
		return s.MapError(w, r, apierr.NewError(404, "not_found", "Secret not found or already viewed"), h.templateService)
	}

	secretResp := SecretResponse{Data: secret.Data}
	return response.JSON(w, 200, secretResp)
}

// RevealOneTimeSecret consumes the one-time secret and returns the content.
// @Summary Reveal one-time secret
// @Description Consumes the one-time secret and returns the content (one-time use only)
// @Tags secret
// @Param short path string true "Short code (6 characters)"
// @Success 200 {string} string "Secret content HTML partial"
// @Failure 404
// @Failure 422
// @Router /api/secrets/{short}/onetime [post]
func (h *SecretHandler) RevealOneTimeSecret(w http.ResponseWriter, r *http.Request) error {
	short := r.PathValue("short")

	if len(short) != 6 {
		return h.templateService.RenderError(w, "Invalid secret code")
	}

	secret, err := h.secretService.GetSecret(short)
	if err != nil {
		return h.templateService.RenderError(w, "This one-time link has already been viewed or does not exist.")
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := map[string]any{
		"Type": "secret",
		"Data": secret.Data,
	}
	return h.templateService.RenderOnetimeReveal(w, data)
}
