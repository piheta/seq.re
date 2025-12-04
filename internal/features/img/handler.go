package img

import (
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/piheta/apicore/apierr"
	"github.com/piheta/apicore/response"
	"github.com/piheta/seq.re/config"
	s "github.com/piheta/seq.re/internal/shared"
)

type ImageHandler struct {
	imageService        *ImageService
	imageViewerTemplate *template.Template
}

func NewImageHandler(imageService *ImageService) *ImageHandler {
	tmpl := template.Must(template.ParseFiles("web/templates/image-viewer.html"))
	return &ImageHandler{
		imageService:        imageService,
		imageViewerTemplate: tmpl,
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
		return apierr.NewError(422, "validation", "Invalid image code")
	}

	imageData, contentType, encrypted, err := h.imageService.GetImage(short)
	if err != nil {
		return apierr.NewError(404, "not_found", "Image not found")
	}

	if s.IsBrowser(r) && encrypted {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := ImageResponse{Data: string(imageData)}
		return h.imageViewerTemplate.Execute(w, data)
	}

	if encrypted {
		return response.JSON(w, 200, ImageResponse{Data: string(imageData)}) // already base64 encoded
	}

	// Otherwise return raw image bytes
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(imageData)
	return nil
}
