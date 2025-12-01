package link

import "time"

type LinkRequest struct {
	URL string `json:"url" validate:"required,notprivateip"`
}

type RedirectRequest struct {
	Short string
}

type Link struct {
	Short     string
	URL       string
	CreatedAt time.Time
	ExpiresAt time.Time
}
