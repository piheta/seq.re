package link

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

type LinkHandler struct {
	linkService      *LinkService
	redirectTemplate *template.Template
}

func NewLinkHandler(linkService *LinkService) *LinkHandler {
	tmpl := template.Must(template.ParseFiles("web/templates/redirect.html"))
	return &LinkHandler{
		linkService:      linkService,
		redirectTemplate: tmpl,
	}
}

// RedirectByShort redirects to the original URL based on a short code.
// For browser requests, serves an HTML page that handles client-side decryption.
// For non-browser requests (CLI, curl), performs server-side redirect.
// @Summary Redirect to original URL
// @Description Redirects to the original URL associated with the given short code
// @Tags link
// @Param short path string true "Short code"
// @Success 301 "Redirect to original URL"
// @Failure 404 "Short code not found"
// @Router /{short} [get]
func (h *LinkHandler) RedirectByShort(w http.ResponseWriter, r *http.Request) error {
	short := r.PathValue("short")

	if len(short) != 6 {
		return apierr.NewError(422, "validation", "Invalid shorturl code")
	}

	link, err := h.linkService.GetLinkByShort(short)
	if err != nil {
		return apierr.NewError(404, "url", "url not found")
	}

	if s.IsBrowser(r) && link.Encrypted {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := LinkResponse{URL: link.URL}
		return h.redirectTemplate.Execute(w, data)
	}

	return response.Redirect(w, r, link.URL)
}

// CreateLink creates a new shortened URL.
// @Summary Create a shortened URL
// @Description Creates a new shortened URL from the provided original URL
// @Tags link
// @Accept json
// @Produce json
// @Param request body LinkRequest true "Link request with URL to shorten"
// @Success 201 {string} string "Shortened URL"
// @Failure 400 "Invalid request or URL format"
// @Failure 500 "Internal server error"
// @Router /api/links [post]
func (h *LinkHandler) CreateLink(w http.ResponseWriter, r *http.Request) error {
	var linkReq LinkRequest
	if err := json.NewDecoder(r.Body).Decode(&linkReq); err != nil {
		return err
	}

	// Skip URL validation for encrypted links (client encrypts before sending)
	// For encrypted links, only verify that URL field is not empty
	if linkReq.Encrypted {
		if linkReq.URL == "" {
			return apierr.NewError(400, "validation", "URL is required")
		}
	} else {
		if err := s.Validate.Struct(&linkReq); err != nil {
			return err
		}
	}

	link, err := h.linkService.CreateLink(linkReq.URL, linkReq.Encrypted, linkReq.OneTime)
	if err != nil {
		return err
	}

	shortURL := fmt.Sprintf("%s%s/%s", config.Config.RedirectHost, config.Config.RedirectPort, link.Short)

	return response.JSON(w, 201, shortURL)
}

// GetLinkByShort retrieves link information by short code.
// @Summary Get link information
// @Description Returns the original URL and expiry time associated with the given short code
// @Tags link
// @Param short path string true "Short code (6 characters)"
// @Success 200 {object} LinkResponse "Link information"
// @Failure 404
// @Failure 422
// @Router /api/links/{short} [get]
func (h *LinkHandler) GetLinkByShort(w http.ResponseWriter, r *http.Request) error {
	short := r.PathValue("short")

	if len(short) != 6 {
		return apierr.NewError(422, "validation", "Invalid shorturl code")
	}

	link, err := h.linkService.GetLinkByShort(short)
	if err != nil {
		return apierr.NewError(404, "url", "url not found")
	}

	linkResp := LinkResponse{
		link.URL,
		link.ExpiresAt,
	}

	return response.JSON(w, 200, linkResp)
}
