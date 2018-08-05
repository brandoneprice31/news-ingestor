package article

import runner "gopkg.in/mgutz/dat.v1/sqlx-runner"

type (
	DB interface {
		FindByURL(url string) (*Article, error)
		FindByURLs(urls []string) ([]Article, error)
		Insert(aa []Article) error
		Update(aa []Article) error
	}

	db struct {
		conn *runner.DB
	}
)

const (
	table = "articles"
)

func NewDB(conn *runner.DB) DB {
	return &db{
		conn: conn,
	}
}

func (db *db) FindByURL(url string) (*Article, error) {
	var a Article
	err := db.conn.
		Select(DBColumns...).
		From(table).
		Where("url=$1", url).
		QueryStruct(&a)

	return &a, err
}

func (db *db) FindByURLs(urls []string) ([]Article, error) {
	var aa []Article
	err := db.conn.
		Select(DBColumns...).
		From(table).
		Where("url in $1", urls).
		QueryStructs(&aa)

	return aa, err
}

func (db *db) Insert(aa []Article) error {
	tx, _ := db.conn.Begin()
	defer tx.AutoRollback()

	for _, a := range aa {
		_, err := tx.
			InsertInto(table).
			Columns(DBColumns...).
			Values(a.Fields()).
			Exec()
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *db) Update(aa []Article) error {
	tx, _ := db.conn.Begin()
	defer tx.AutoRollback()

	for _, a := range aa {
		_, err := tx.
			Update(table).
			Set("source", a.Source).
			Set("title", a.Title).
			Set("author", a.Author).
			Set("date", a.Date).
			Set("text", a.Text).
			Set("headline", a.Headline).
			Set("url", a.URL).
			Where("url=$1", a.URL).
			Exec()
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
