package secret

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/piheta/apicore/apierr"
	"github.com/piheta/apicore/response"
	"github.com/piheta/seq.re/config"
	s "github.com/piheta/seq.re/internal/shared"
)

type SecretHandler struct {
	secretService        *SecretService
	secretViewerTemplate *template.Template
	resultTemplate       *template.Template
}

func NewSecretHandler(secretService *SecretService) *SecretHandler {
	viewerTmpl := template.Must(template.ParseFiles("web/templates/secret-viewer.html"))
	resultTmpl := template.Must(template.ParseFiles("web/templates/partials/generic-result.html"))
	return &SecretHandler{
		secretService:        secretService,
		secretViewerTemplate: viewerTmpl,
		resultTemplate:       resultTmpl,
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
		return h.resultTemplate.Execute(w, data)
	}

	return response.JSON(w, 201, secretURL)
}

// GetSecretByShort retrieves secret information by short code.
// @Summary Get secret information
// @Description Returns the original URL and expiry time associated with the given short code
// @Tags secret
// @Param short path string true "Short code (6 characters)"
// @Success 200 {object} SecretResponse "Secret information"
// @Failure 404
// @Failure 422
// @Router /s/{short} [get]
func (h *SecretHandler) GetSecretByShort(w http.ResponseWriter, r *http.Request) error {
	short := r.PathValue("short")

	if len(short) != 6 {
		return apierr.NewError(422, "validation", "Invalid shorturl code")
	}

	secret, err := h.secretService.GetSecret(short)
	if err != nil {
		return apierr.NewError(404, "not_found", "secret not found or already viewed")
	}

	if s.IsBrowser(r) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := SecretResponse{Data: secret.Data}
		return h.secretViewerTemplate.Execute(w, data)
	}

	secretResp := SecretResponse{Data: secret.Data}

	return response.JSON(w, 200, secretResp)
}
