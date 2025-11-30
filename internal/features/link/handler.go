package link

import (
	"encoding/json"
	"fmt"
	"net/http"

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

func (h *LinkHandler) RedirectByShort(w http.ResponseWriter, r *http.Request) error {
	short := r.PathValue("short")

	link, err := h.linkService.GetLinkByShort(short)
	if err != nil {
		return err
	}

	http.Redirect(w, r, link.URL, http.StatusMovedPermanently)

	return nil
}

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

	shortURL := fmt.Sprintf("%s%s/%s", config.Config.Host, config.Config.Port, link.Short)

	return response.JSON(w, 201, shortURL)
}
