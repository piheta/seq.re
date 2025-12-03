package link

import "time"

type LinkRequest struct {
	URL       string `json:"url" validate:"required,notprivateip"`
	Encrypted bool   `json:"encrypted"`
	OneTime   bool   `json:"onetime"`
}

type LinkResponse struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

type RedirectRequest struct {
	Short string
}

type Link struct {
	Short     string
	URL       string
	Encrypted bool
	OneTime   bool
	CreatedAt time.Time
	ExpiresAt time.Time
}
