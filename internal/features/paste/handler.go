package paste

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/piheta/apicore/apierr"
	"github.com/piheta/apicore/response"
	"github.com/piheta/seq.re/config"
	"github.com/piheta/seq.re/internal/shared"
)

type PasteHandler struct {
	pasteService        *PasteService
	pasteViewerTemplate *template.Template
	resultTemplate      *template.Template
}

func NewPasteHandler(pasteService *PasteService) *PasteHandler {
	viewerTmpl := template.Must(template.ParseFiles("web/templates/paste-viewer.html"))
	resultTmpl := template.Must(template.ParseFiles("web/templates/partials/generic-result.html"))
	return &PasteHandler{
		pasteService:        pasteService,
		pasteViewerTemplate: viewerTmpl,
		resultTemplate:      resultTmpl,
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
		return h.resultTemplate.Execute(w, data)
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
		return apierr.NewError(422, "validation", "Invalid paste code")
	}

	paste, err := h.pasteService.GetPaste(short)
	if err != nil {
		return apierr.NewError(404, "not_found", "Paste not found")
	}

	if shared.IsBrowser(r) && paste.Encrypted {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := struct {
			Data     string
			Language string
		}{
			Data:     paste.Content,
			Language: paste.Language,
		}
		return h.pasteViewerTemplate.Execute(w, data)
	}

	if paste.Encrypted {
		return response.JSON(w, 200, PasteResponse{Data: paste.Content}) // already base64 encoded
	}

	// Return as plain text with UTF-8 charset
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(paste.Content))
	return nil
}
