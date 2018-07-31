package article

import runner "gopkg.in/mgutz/dat.v1/sqlx-runner"

type (
	DB interface {
		FindByURL(url string) (*Article, error)
		Insert(aa []Article) error
	}

	db struct {
		conn *runner.DB
	}
)

func NewDB(conn *runner.DB) DB {
	return &db{
		conn: conn,
	}
}

func (db *db) FindByURL(url string) (*Article, error) {
	var a Article
	err := db.conn.
		Select(columns...).
		From(table).
		Where("url=$1", url).
		QueryStruct(&a)

	return &a, err
}

func (db *db) Insert(aa []Article) error {
	tx, _ := db.conn.Begin()
	defer tx.AutoRollback()

	for _, a := range aa {
		_, err := tx.
			InsertInto(table).
			Columns(columns...).
			Values(a.Source, a.Title, a.Author, a.Date, a.Text, a.Headline, a.URL).
			Exec()
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
