package img

import "time"

type Image struct {
	Short       string
	FilePath    string
	ContentType string
	Encrypted   bool
	OneTime     bool
	CreatedAt   time.Time
	ExpiresAt   time.Time
}

type ImageResponse struct {
	Data string `json:"data"`
}
