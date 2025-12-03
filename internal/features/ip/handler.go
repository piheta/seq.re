package ip

import (
	"github.com/piheta/apicore/response"
	"net/http"
)

type IPHandler struct {
	ipService *IPService
}

func NewIPHandler(ipService *IPService) *IPHandler {
	return &IPHandler{ipService: ipService}
}

// GetPublicIP retrieves the client's public IP
// @Summary Get client public IP
// @Description Returns the public IP of the client making the request
// @Tags ip
// @Accept json
// @Produce json
// @Success 200 {object} IP "IP retrieved successfully"
// @Router /api/ip [get]
func (h *IPHandler) GetPublicIP(w http.ResponseWriter, r *http.Request) error {
	ip := h.ipService.GetClientIP(r)

	return response.JSON(w, 200, ip)
}
