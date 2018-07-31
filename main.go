package main

import (
	"database/sql"
	"os"

	"github.com/brandoneprice31/news-ingestor/article"
	"github.com/brandoneprice31/news-ingestor/config"
	"github.com/brandoneprice31/news-ingestor/ingestor"
	dat "gopkg.in/mgutz/dat.v1"
	runner "gopkg.in/mgutz/dat.v1/sqlx-runner"

	"github.com/rs/zerolog"
)

const (
	IngestionRounds   = 1
	ArticleBufferSize = 1
)

var (
	articleService article.Service
	log            = zerolog.New(os.Stderr).With().Timestamp().Logger()
	ingestors      = []ingestor.Ingestor{ingestor.FoxIngestor()}
)

func main() {
	c, err := config.Load(config.GetEnv())
	if err != nil {
		panic(err)
	}

	// Connect to postgres db.
	sqlDB, err := sql.Open("postgres", c.PostgresURL())
	defer sqlDB.Close()
	if err != nil {
		panic(err)
	}
	db := runner.NewDB(sqlDB, "postgres")
	dat.EnableInterpolation = true

	// create services
	articleService = article.NewService(db)

	// begin ingesting
	articleBuffer := make(chan article.Article, ArticleBufferSize)
	startIngestion(articleBuffer)
	saveArticles(articleBuffer)
}

// ingests articles from the slice of ingestors and appends them to the channel
func startIngestion(articleBuffer chan article.Article) {
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
}

// saves the articles inside the channel in our db
func saveArticles(articleBuffer chan article.Article) {
	for {
		// save articles in db
		aa := make([]article.Article, ArticleBufferSize)
		for i := range aa {
			aa[i] = <-articleBuffer
		}

		log.Info().Msgf("attemping to save %d articles", len(aa))

		n, err := articleService.Save(aa)
		if err != nil {
			log.Error().Err(err)
			return
		}

		log.Info().Msgf("saved %d articles", n)
	}
}
