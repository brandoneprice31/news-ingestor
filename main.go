package main

import (
	"database/sql"
	"os"
	"time"

	"github.com/brandoneprice31/news-ingestor/article"
	"github.com/brandoneprice31/news-ingestor/config"
	"github.com/brandoneprice31/news-ingestor/ingestor"
	dat "gopkg.in/mgutz/dat.v1"
	runner "gopkg.in/mgutz/dat.v1/sqlx-runner"

	"github.com/rs/zerolog"
)

const (
	IngestionRounds   = 1
	ArticleBufferSize = 10
)

var (
	articleService article.Service
	log            = zerolog.New(os.Stderr).With().Timestamp().Logger()
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
	articleBuffer := make(chan articleBool, ArticleBufferSize)
	startIngestion(articleBuffer)
	saveArticles(articleBuffer)
}

type articleBool struct {
	a article.Article
	b bool
}

// ingests articles from the slice of ingestors and appends them to the channel
func startIngestion(articleBuffer chan articleBool) {
	// spin up ingestors
	for iter := range ingestor.Ingestors {
		i := ingestor.Ingestors[iter]

		go func() {
			for round := 0; round < IngestionRounds; round++ {
				log.Info().Str("source", i.Source()).Msgf("starting ingestion")

				aa, err := i.Ingest()
				if err != nil {
					log.Fatal().Err(err)
				}

				log.Info().Str("source", i.Source()).Msgf("completed ingestion")

				for _, a := range aa {
					articleBuffer <- articleBool{a: a, b: true}
				}
			}
		}()
	}
}

// saves the articles inside the channel in our db
func saveArticles(articleBuffer chan articleBool) {
	for {
		breakOut := false
		aa := []article.Article{}
		go func() {
			for {
				time.Sleep(5 * time.Second)
				if len(aa) < ArticleBufferSize {
					articleBuffer <- articleBool{b: false}
					break
				}
			}
		}()

		for i := 0; i < ArticleBufferSize; i++ {
			ab := <-articleBuffer
			if !ab.b {
				breakOut = true
				break
			}
			aa = append(aa, ab.a)
		}

		log.Info().Msgf("attemping to save %d articles", len(aa))

		inserted, updated, err := articleService.Save(aa)
		if err != nil {
			log.Error().Err(err)
			return
		}

		log.Info().Msgf("inserted %d articles", inserted)
		log.Info().Msgf("updated %d articles", updated)

		if breakOut {
			break
		}
	}
}
