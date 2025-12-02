package secret

import "time"

type SecretRequest struct {
	Data string `json:"data" validate:"required,base64,min=44"`
}

type SecretResponse struct {
	Data string `json:"data"`
}

type Secret struct {
	Short     string
	Data      string
	CreatedAt time.Time
	ExpiresAt time.Time
}
