package link

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/piheta/apicore/apierr"
	"github.com/piheta/apicore/response"
	"github.com/piheta/seq.re/config"
	s "github.com/piheta/seq.re/internal/shared"
)

type LinkHandler struct {
	linkService     *LinkService
	templateService *s.TemplateService
}

func NewLinkHandler(linkService *LinkService, templateService *s.TemplateService) *LinkHandler {
	return &LinkHandler{
		linkService:     linkService,
		templateService: templateService,
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
		return s.MapError(w, r, apierr.NewError(422, "validation", "Invalid link code"), h.templateService)
	}

	link, err := h.linkService.CheckLinkExists(short)
	if err != nil {
		return s.MapError(w, r, apierr.NewError(404, "url", "Link not found"), h.templateService)
	}

	if link.OneTime && r.URL.Query().Get("cli") != "true" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := map[string]string{
			"ID":   short,
			"Type": "url",
		}
		return h.templateService.RenderOnetime(w, data)
	}

	link, err = h.linkService.GetLinkByShort(short)
	if err != nil {
		return s.MapError(w, r, apierr.NewError(404, "url", "Link not found"), h.templateService)
	}

	if r.URL.Query().Get("cli") != "true" && link.Encrypted {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := LinkResponse{URL: link.URL}
		return h.templateService.RenderRedirect(w, data)
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
		if !strings.Contains(linkReq.URL, "://") {
			linkReq.URL = "https://" + linkReq.URL
		}
		if err := s.Validate.Struct(&linkReq); err != nil {
			return err
		}
	}

	link, err := h.linkService.CreateLink(linkReq.URL, linkReq.Encrypted, linkReq.OneTime)
	if err != nil {
		return err
	}

	shortURL := fmt.Sprintf("%s%s/%s", config.Config.RedirectHost, config.Config.RedirectPort, link.Short)

	// Check if this is an HTMX request (wants HTML response)
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := map[string]string{
			"URL":      shortURL,
			"ButtonID": "url",
		}
		if linkReq.OneTime {
			data["Warning"] = "This link will self-destruct after being viewed once."
		}
		return h.templateService.RenderResult(w, data)
	}

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
		return s.MapError(w, r, apierr.NewError(422, "validation", "Invalid link code"), h.templateService)
	}

	link, err := h.linkService.GetLinkByShort(short)
	if err != nil {
		return s.MapError(w, r, apierr.NewError(404, "url", "Link not found"), h.templateService)
	}

	linkResp := LinkResponse{
		link.URL,
		link.ExpiresAt,
	}

	return response.JSON(w, 200, linkResp)
}

// RevealOneTimeLink consumes the one-time link and returns the content for redirect.
// @Summary Reveal one-time link
// @Description Consumes the one-time link and returns the URL for redirect (one-time use only)
// @Tags link
// @Param short path string true "Short code (6 characters)"
// @Success 200 {string} string "Link content HTML partial"
// @Failure 404
// @Failure 422
// @Router /api/links/{short}/onetime [post]
func (h *LinkHandler) RevealOneTimeLink(w http.ResponseWriter, r *http.Request) error {
	short := r.PathValue("short")

	if len(short) != 6 {
		return h.templateService.RenderError(w, "Invalid link code")
	}

	link, err := h.linkService.GetLinkByShort(short)
	if err != nil {
		return h.templateService.RenderError(w, "This one-time link has already been viewed or does not exist.")
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := map[string]any{
		"Type": "url",
		"Data": link.URL,
	}

	return h.templateService.RenderOnetimeReveal(w, data)
}
