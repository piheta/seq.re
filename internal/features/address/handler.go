package address

import (
	"github.com/piheta/apicore/response"
	"net/http"
)

type AddressHandler struct {
	addressService *AddressService
}

func NewAddressHandler(addressService *AddressService) *AddressHandler {
	return &AddressHandler{addressService: addressService}
}

// GetPublicIP retrieves the client's public IP address.
// @Summary Get client public IP
// @Description Returns the public IP address of the client making the request
// @Tags address
// @Accept json
// @Produce json
// @Success 200 {object} Address "IP address retrieved successfully"
// @Router /api/ip [get]
func (h *AddressHandler) GetPublicIP(w http.ResponseWriter, r *http.Request) error {
	ip := h.addressService.GetClientIP(r)

	return response.JSON(w, 200, ip)
}
