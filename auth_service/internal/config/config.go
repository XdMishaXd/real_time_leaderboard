package config

import (
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `yaml:"env" env-default:"local"`
	Tokens     `yaml:"tokens"`
	RabbitMQ   `yaml:"rabbitmq"`
	Postgres   `yaml:"postgres"`
	HTTPServer `yaml:"http_server"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type Postgres struct {
	Host     string `yaml:"host" env-default:"postgres"`
	Port     int    `yaml:"port" env-default:"5432"`
	User     string `yaml:"user" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
	DBName   string `yaml:"dbname" env-required:"true"`
	SSLMode  string `yaml:"sslmode" env-default:"disabled"`
}

type Tokens struct {
	AccessTokenTTL          time.Duration `yaml:"access_token_ttl" env-required:"true"`
	RefreshTokenTTL         time.Duration `yaml:"refresh_token_ttl" env-required:"true"`
	VerificationTokenTTL    time.Duration `yaml:"verification_token_ttl" env-required:"true"`
	VerificationTokenSecret string        `yaml:"verification_token_secret" env-required:"true"`
}

type RabbitMQ struct {
	URL       string `yaml:"url" env-required:"true"`
	QueueName string `yaml:"queue_name" env-required:"true"`
}

func MustLoad(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("Config file does not exist" + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("Failed to read config" + err.Error())
	}

	return &cfg
}
