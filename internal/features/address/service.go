package address

import (
	"net/http"

	"github.com/piheta/seq.re/internal/shared"
)

type AddressService struct {
}

func NewAddressService() *AddressService {
	return &AddressService{}
}

func (s *AddressService) GetClientIP(r *http.Request) Address {
	return Address{IP: shared.GetIP(r)}
}
