package config

import (
	"log"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Env  string `env:"ENV" envDefault:"local"`
	HTTP HTTPServer
	DB   Postgres
}

type HTTPServer struct {
	Port    string        `env:"HTTP_PORT" envDefault:"8080"`
	Timeout time.Duration `env:"HTTP_TIMEOUT" envDefault:"5s"`
}

type Postgres struct {
	Host     string `env:"DB_HOST,required"`
	Port     string `env:"DB_PORT,required"`
	User     string `env:"POSTGRES_USER,required"`     // Match .env
	Password string `env:"POSTGRES_PASSWORD,required"` // Match .env
	Name     string `env:"POSTGRES_DB,required"`       // Match .env
}

func MustLoad() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("unable to parse config: %v", err)
	}

	return &cfg
}
