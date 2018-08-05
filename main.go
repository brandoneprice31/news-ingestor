package main

import (
	"database/sql"
	"os"
	"sync"

	"github.com/brandoneprice31/news-ingestor/article"
	"github.com/brandoneprice31/news-ingestor/config"
	"github.com/brandoneprice31/news-ingestor/ingestor"
	dat "gopkg.in/mgutz/dat.v1"
	runner "gopkg.in/mgutz/dat.v1/sqlx-runner"

	"github.com/rs/zerolog"
)

const (
	IngestionRounds   = 1
	ArticleBufferSize = 100
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
	articleBuffer, wg := make(chan articleBool, ArticleBufferSize), sync.WaitGroup{}
	startIngestion(articleBuffer, &wg)
	saveArticles(articleBuffer, &wg)
}

type articleBool struct {
	a article.Article
	b bool
}

// ingests articles from the slice of ingestors and appends them to the channel
func startIngestion(articleBuffer chan articleBool, wg *sync.WaitGroup) {
	// spin up ingestors
	for iter := range ingestor.Ingestors {
		i := ingestor.Ingestors[iter]

		wg.Add(1)
		go func() {
			for round := 0; round < IngestionRounds; round++ {
				defer wg.Done()
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
func saveArticles(articleBuffer chan articleBool, wg *sync.WaitGroup) {
	aa := []article.Article{}

	// spin up routine that will force flush the buffer once the waitgroup is done
	go func() {
		wg.Wait()
		articleBuffer <- articleBool{b: false}
	}()

	// continuously save the buffer
	breakout := false
	for !breakout {
		for i := 0; i < ArticleBufferSize; i++ {
			ab := <-articleBuffer
			if !ab.b {
				breakout = true
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

		aa = []article.Article{}
	}
}
