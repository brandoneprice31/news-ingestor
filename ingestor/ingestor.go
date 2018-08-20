package ingestor

import (
	"github.com/brandoneprice31/news-ingestor/article"
)

type (
	Ingestor interface {
		Source() string
		Ingest() ([]article.Article, error)
	}
)

var Ingestors = []Ingestor{Fox(), Breitbart(), NYT()}
