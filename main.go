package main

import (
	"os"

	"github.com/brandoneprice31/news-ingestor/article"
	"github.com/brandoneprice31/news-ingestor/ingestor"

	"github.com/rs/zerolog"
)

const (
	IngestionRounds = 1
)

var (
	log       = zerolog.New(os.Stderr).With().Timestamp().Logger()
	ingestors = []ingestor.Ingestor{ingestor.FoxIngestor()}
)

func main() {
	articleBuffer := make(chan article.Article)

	// spin up ingestors
	for iter := range ingestors {
		i := ingestors[iter]

		go func() {
			for round := 0; round < IngestionRounds; round++ {
				log.Info().Str("source", i.Source()).Msgf("starting ingestion")

				aa, err := i.Ingest()
				if err != nil {
					log.Fatal().Err(err)
				}

				log.Info().Str("source", i.Source()).Msgf("completed ingestion")

				for _, a := range aa {
					articleBuffer <- a
				}
			}
		}()
	}

	// save articles in db
	for a := range articleBuffer {
		save(a)
	}
}

func save(a article.Article) {
	log.Info().Msgf("title: %s, author: %s", a.Title, a.Author)
}
