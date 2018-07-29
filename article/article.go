package article

import "time"

type (
	Article struct {
		Title    string    `json:"title"`
		Author   string    `json:"author"`
		Date     time.Time `json:"date"`
		Text     string    `json:"text"`
		Headline bool      `json:"headline"`
	}
)
