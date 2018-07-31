package article

import (
	"database/sql"

	runner "gopkg.in/mgutz/dat.v1/sqlx-runner"
)

type (
	Service interface {
		Save([]Article) (int, error)
	}

	service struct {
		db DB
	}
)

func NewService(conn *runner.DB) Service {
	return &service{db: NewDB(conn)}
}

func (s *service) Save(aa []Article) (int, error) {
	// clean the articles
	urlMap, cleanedArticles := make(map[string]bool), []Article{}
	for _, a := range aa {
		if urlMap[a.URL] {
			continue
		}

		_, err := s.db.FindByURL(a.URL)
		if err != sql.ErrNoRows {
			continue
		}

		cleanedArticles = append(cleanedArticles, a)
		urlMap[a.URL] = true
	}

	// save them in the db
	return len(cleanedArticles), s.db.Insert(cleanedArticles)
}
