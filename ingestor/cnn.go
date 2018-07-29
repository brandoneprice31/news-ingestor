package ingestor

import (
	"github.com/brandoneprice31/news-ingestor/article"
)

type (
	cnn struct {
		host string
	}
)

func CNNIngestor() Ingestor {

	return cnn{
		host: "www.cnn.com",
	}
}

func (c cnn) Source() string {
	return "cnn"
}

func (c cnn) Ingest() ([]article.Article, error) {
	return nil, nil
}
