package paste

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/piheta/apicore/apierr"
	"github.com/piheta/apicore/response"
	"github.com/piheta/seq.re/config"
	"github.com/piheta/seq.re/internal/shared"
)

type PasteHandler struct {
	pasteService    *PasteService
	templateService *shared.TemplateService
}

func NewPasteHandler(pasteService *PasteService, templateService *shared.TemplateService) *PasteHandler {
	return &PasteHandler{
		pasteService:    pasteService,
		templateService: templateService,
	}
}

// CreatePaste creates a new text paste
// @Summary Create a paste
// @Description Creates a text paste (code, logs, plain text)
// @Tags paste
// @Accept json
// @Produce json
// @Param paste body CreatePasteRequest true "Paste content and options"
// @Success 201 {string} string "Paste URL"
// @Failure 400 "Invalid request"
// @Failure 500 "Internal server error"
// @Router /api/pastes [post]
func (h *PasteHandler) CreatePaste(w http.ResponseWriter, r *http.Request) error {
	var req CreatePasteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierr.NewError(400, "invalid_request", "Failed to parse request body")
	}

	if err := shared.Validate.Struct(req); err != nil {
		return apierr.NewError(400, "validation", err.Error())
	}

	paste, err := h.pasteService.CreatePaste(req.Content, req.Language, req.Encrypted, req.OneTime)
	if err != nil {
		return err
	}

	pasteURL := fmt.Sprintf("%s%s/p/%s", config.Config.RedirectHost, config.Config.RedirectPort, paste.Short)

	// Check if this is an HTMX request (wants HTML response)
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := map[string]string{
			"URL":      pasteURL,
			"ButtonID": "code",
		}
		if req.OneTime {
			data["Warning"] = "This link will self-destruct after being viewed once."
		}
		return h.templateService.RenderResult(w, data)
	}

	return response.JSON(w, 201, pasteURL)
}

// GetPasteByShort retrieves and serves the paste content
// @Summary Get paste by short code
// @Description Returns the paste content as text/plain (or encrypted data as JSON if encrypted)
// @Tags paste
// @Param short path string true "Short code (6 characters)"
// @Success 200 {string} string "Paste content"
// @Failure 404
// @Failure 422
// @Router /p/{short} [get]
func (h *PasteHandler) GetPasteByShort(w http.ResponseWriter, r *http.Request) error {
	short := r.PathValue("short")

	if len(short) != 6 {
		return shared.MapError(w, r, apierr.NewError(422, "validation", "Invalid paste code"), h.templateService)
	}

	paste, err := h.pasteService.CheckPasteExists(short)
	if err != nil {
		return shared.MapError(w, r, apierr.NewError(404, "not_found", "Paste not found"), h.templateService)
	}

	if paste.OneTime && r.URL.Query().Get("cli") != "true" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := map[string]string{
			"ID":   short,
			"Type": "code",
		}
		return h.templateService.RenderOnetime(w, data)
	}

	paste, err = h.pasteService.GetPaste(short)
	if err != nil {
		return shared.MapError(w, r, apierr.NewError(404, "not_found", "Paste not found"), h.templateService)
	}

	if r.URL.Query().Get("cli") != "true" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := map[string]any{
			"Type":      "code",
			"Data":      paste.Content,
			"Encrypted": paste.Encrypted,
			"Metadata": map[string]string{
				"Language": paste.Language,
			},
		}
		return h.templateService.RenderContentViewer(w, data)
	}

	if paste.Encrypted {
		return response.JSON(w, 200, PasteResponse{Data: paste.Content})
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(paste.Content))
	return nil
}

// RevealOneTimePaste consumes the one-time paste and returns the content.
// @Summary Reveal one-time paste
// @Description Consumes the one-time paste and returns the content (one-time use only)
// @Tags paste
// @Param short path string true "Short code (6 characters)"
// @Success 200 {string} string "Paste content HTML partial"
// @Failure 404
// @Failure 422
// @Router /api/pastes/{short}/onetime [post]
func (h *PasteHandler) RevealOneTimePaste(w http.ResponseWriter, r *http.Request) error {
	short := r.PathValue("short")

	if len(short) != 6 {
		return h.templateService.RenderError(w, "Invalid paste code")
	}

	paste, err := h.pasteService.GetPaste(short)
	if err != nil {
		return h.templateService.RenderError(w, "This one-time link has already been viewed or does not exist.")
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := map[string]any{
		"Type": "code",
		"Data": paste.Content,
		"Metadata": map[string]string{
			"Language": paste.Language,
		},
	}

	return h.templateService.RenderOnetimeReveal(w, data)
}
