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

var (
	DBColumns = []string{"source", "title", "author", "date", "text", "headline", "url"}
)

// Returns each field in the same order as DBColumns
func (a *Article) Fields() (string, string, string, time.Time, string, bool, string) {
	return a.Source, a.Title, a.Author, a.Date, a.Text, a.Headline, a.URL
}
