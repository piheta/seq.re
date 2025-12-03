package ip

import (
	"net/http"

	"github.com/piheta/seq.re/internal/shared"
)

type IPService struct {
}

func NewIPService() *IPService {
	return &IPService{}
}

func (s *IPService) GetClientIP(r *http.Request) IP {
	return IP{IP: shared.GetIP(r)}
}
