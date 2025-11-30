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

func (h *AddressHandler) GetPublicIP(w http.ResponseWriter, r *http.Request) error {
	ip := h.addressService.GetClientIP(r)

	return response.JSON(w, 200, ip)
}
