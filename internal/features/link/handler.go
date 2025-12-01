package link

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/piheta/apicore/apierr"
	"github.com/piheta/apicore/response"
	"github.com/piheta/seq.re/config"
	s "github.com/piheta/seq.re/internal/shared"
)

type LinkHandler struct {
	linkService *LinkService
}

func NewLinkHandler(linkService *LinkService) *LinkHandler {
	return &LinkHandler{linkService: linkService}
}

// RedirectByShort redirects to the original URL based on a short code.
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
// @Router /api/link [post]
func (h *LinkHandler) CreateLink(w http.ResponseWriter, r *http.Request) error {
	var linkReq LinkRequest
	if err := json.NewDecoder(r.Body).Decode(&linkReq); err != nil {
		return err
	}

	if err := s.Validate.Struct(&linkReq); err != nil {
		return err
	}

	link, err := h.linkService.CreateLink(linkReq.URL)
	if err != nil {
		return err
	}

	shortURL := fmt.Sprintf("%s%s/%s", config.Config.RedirectHost, config.Config.RedirectPort, link.Short)

	return response.JSON(w, 201, shortURL)
}
