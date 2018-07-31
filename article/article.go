package article

import "time"

type (
	Article struct {
		Source   string    `json:"source" db:"source"`
		Title    string    `json:"title" db:"title"`
		Author   string    `json:"author" db:"author"`
		Date     time.Time `json:"date" db:"date"`
		Text     string    `json:"text" db:"text"`
		Headline bool      `json:"headline" db:"headline"`
		URL      string    `json:"url" db:"url"`
	}
)
