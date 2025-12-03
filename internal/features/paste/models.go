package paste

import "time"

type Paste struct {
	Short     string
	Content   string
	Language  string // Optional: "go", "python", "json", "markdown", etc.
	Encrypted bool
	OneTime   bool
	CreatedAt time.Time
	ExpiresAt time.Time
}

type PasteResponse struct {
	Data string `json:"data"`
}

type CreatePasteRequest struct {
	Content   string `json:"content" validate:"required,max=1048576"` // 1MB max
	Language  string `json:"language,omitempty" validate:"omitempty,max=50"`
	Encrypted bool   `json:"encrypted"`
	OneTime   bool   `json:"onetime"`
}
