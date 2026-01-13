package shared // nolint

import (
	"fmt"
	"net/http"

	"github.com/piheta/apicore/apierr"
)

// MapError handles error responses based on the request type (browser vs API).
// For browser requests, it renders a user-friendly HTML error page.
// For API requests, it returns the structured API error.
func MapError(w http.ResponseWriter, r *http.Request, err error, templateService *TemplateService) error {
	apiErr, ok := err.(*apierr.APIError)
	if !ok {
		return err
	}

	if r.URL.Query().Get("cli") != "true" {
		message := fmt.Sprintf("%v", apiErr.Message)
		return templateService.RenderError(w, message)
	}

	return apiErr
}
