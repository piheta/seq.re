package img

import (
	"fmt"
	"io"
	"net/http"

	"github.com/piheta/apicore/apierr"
	"github.com/piheta/apicore/response"
	"github.com/piheta/seq.re/config"
	s "github.com/piheta/seq.re/internal/shared"
)

type ImageHandler struct {
	imageService    *ImageService
	templateService *s.TemplateService
}

func NewImageHandler(imageService *ImageService, templateService *s.TemplateService) *ImageHandler {
	return &ImageHandler{
		imageService:    imageService,
		templateService: templateService,
	}
}

// CreateImage uploads an image file
// @Summary Upload an image
// @Description Uploads an image file (raw or encrypted blob)
// @Tags image
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Image file to upload"
// @Param encrypted formData bool false "Whether the file is encrypted"
// @Success 201 {string} string "Image URL"
// @Failure 400 "Invalid request"
// @Failure 500 "Internal server error"
// @Router /api/images [post]
func (h *ImageHandler) CreateImage(w http.ResponseWriter, r *http.Request) error {
	// Parse multipart form (32MB max)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return apierr.NewError(400, "invalid_request", "Failed to parse multipart form")
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return apierr.NewError(400, "invalid_request", "No file provided")
	}
	defer func() {
		_ = file.Close()
	}()

	fileData, err := io.ReadAll(file)
	if err != nil {
		return apierr.NewError(500, "read_error", "Failed to read file")
	}

	encrypted := r.FormValue("encrypted") == "true"
	onetime := r.FormValue("onetime") == "true"

	contentType := http.DetectContentType(fileData)
	if contentType == "application/octet-stream" {
		headerType := header.Header.Get("Content-Type")
		if headerType != "" && headerType != "application/octet-stream" {
			contentType = headerType
		}
	}

	image, err := h.imageService.CreateImage(fileData, contentType, encrypted, onetime)
	if err != nil {
		return err
	}

	imageURL := fmt.Sprintf("%s%s/i/%s", config.Config.RedirectHost, config.Config.RedirectPort, image.Short)

	// Check if this is an HTMX request (wants HTML response)
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := map[string]string{
			"URL":      imageURL,
			"ButtonID": "image",
		}
		if onetime {
			data["Warning"] = "This link will self-destruct after being viewed once."
		}
		return h.templateService.RenderResult(w, data)
	}

	return response.JSON(w, 201, imageURL)
}

// GetImageByShort retrieves and serves the image file
// @Summary Get image by short code
// @Description Returns the raw image file for the given short code (or encrypted data as JSON if encrypted)
// @Tags image
// @Param short path string true "Short code (6 characters)"
// @Success 200 {file} binary "Image file"
// @Failure 404
// @Failure 422
// @Router /i/{short} [get]
func (h *ImageHandler) GetImageByShort(w http.ResponseWriter, r *http.Request) error {
	short := r.PathValue("short")

	if len(short) != 6 {
		return s.MapError(w, r, apierr.NewError(422, "validation", "Invalid image code"), h.templateService)
	}

	imageCheck, err := h.imageService.CheckImageExists(short)
	if err != nil {
		return s.MapError(w, r, apierr.NewError(404, "not_found", "Image not found"), h.templateService)
	}

	if imageCheck.OneTime && r.URL.Query().Get("cli") != "true" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := map[string]string{
			"ID":   short,
			"Type": "image",
		}
		return h.templateService.RenderOnetime(w, data)
	}

	image, imageData, err := h.imageService.GetImage(short)
	if err != nil {
		return s.MapError(w, r, apierr.NewError(404, "not_found", "Image not found"), h.templateService)
	}

	if image.Encrypted {
		if r.URL.Query().Get("cli") != "true" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			data := map[string]string{
				"Data":        string(imageData),
				"ContentType": image.ContentType,
			}
			return h.templateService.RenderImageDecrypt(w, data)
		}
		return response.JSON(w, 200, ImageResponse{Data: string(imageData)})
	}

	w.Header().Set("Content-Type", image.ContentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(imageData)
	return nil
}

// RevealOneTimeImage consumes the one-time image and returns the raw image data.
// @Summary Reveal one-time image
// @Description Consumes the one-time image and returns the raw image (one-time use only)
// @Tags image
// @Param short path string true "Short code (6 characters)"
// @Success 200 {file} binary "Image file"
// @Failure 404
// @Failure 422
// @Router /api/images/{short}/onetime [post]
func (h *ImageHandler) RevealOneTimeImage(w http.ResponseWriter, r *http.Request) error {
	short := r.PathValue("short")

	if len(short) != 6 {
		return h.templateService.RenderError(w, "Invalid image code")
	}

	image, imageData, err := h.imageService.GetImage(short)
	if err != nil {
		return h.templateService.RenderError(w, "This one-time link has already been viewed or does not exist.")
	}

	if image.Encrypted {
		w.Header().Set("Content-Type", "application/json")
		return response.JSON(w, 200, ImageResponse{
			Data: string(imageData), // already base64 encoded
		})
	}

	w.Header().Set("Content-Type", image.ContentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(imageData)
	return nil
}
