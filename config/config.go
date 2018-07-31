package config

import (
	"fmt"
	"os"

	"github.com/vrischmann/envconfig"
)

type Config struct {
	Environment string `envconfig:"-"`
	Postgres    Postgres
}

type Postgres struct {
	Username string `envconfig:"default=brandonprice"`
	Password string `envconfig:"default=password"`
	Host     string `envconfig:"-"`
	DB       string `envconfig:"default=news_ingestor"`
	Params   string `envconfig:"-"`
}

var configs = map[string]Config{
	"development": Config{
		Environment: "development",
		Postgres: Postgres{
			Host:   "localhost:5432",
			Params: "sslmode=disable",
		},
	},
}

func (c *Config) PostgresURL() string {
	pg := c.Postgres
	url := fmt.Sprintf("postgres://%s:%s@%s/%s", pg.Username, pg.Password, pg.Host, pg.DB)
	if pg.Params != "" {
		url = fmt.Sprintf("%s?%s", url, pg.Params)
	}
	return url
}

func Load(env string) (Config, error) {
	config, ok := configs[env]
	if !ok {
		return config, fmt.Errorf("Unknown environment: %s", env)
	}

	err := envconfig.Init(&config)

	return config, err
}

func GetEnv() string {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	return env
}
