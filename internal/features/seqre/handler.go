package seqre

import (
	"github.com/piheta/apicore/response"
	"net/http"
)

type SeqreHandler struct {
	version string
	commit  string
	date    string
}

func NewSeqreHandler(version, commit, date string) *SeqreHandler {
	return &SeqreHandler{version: version, commit: commit, date: date}
}

// GetVersion retrieves the seqre server version
// @Summary Get Seqre server version
// @Description Returns the version of the seqre server
// @Tags seqre
// @Accept json
// @Produce json
// @Success 200 {string}
// @Router /api/version [get]
func (h *SeqreHandler) GetVersion(w http.ResponseWriter, _ *http.Request) error {
	version := VersionResponse{
		Version: h.version,
		Commit:  h.commit,
		Date:    h.date,
	}
	return response.JSON(w, 200, version)
}
