package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type (
	Config struct {
		Env       string          `yaml:"env" env-default:"local"`
		Postgres  PostgresConfig  `yaml:"postgres"`
		Redis     RedisConfig     `yaml:"redis"`
		HTTP      HTTPConfig      `yaml:"http"`
		MusicInfo MusicInfoConfig `yaml:"music_info"`
	}

	PostgresConfig struct {
		Address  string `yaml:"address" env-required:"true"`
		User     string `yaml:"user" env-required:"true"`
		Password string `yaml:"password" env-required:"true" env:"POSTGRES_PASSWORD"`
		DBName   string `yaml:"dbname" env-required:"true"`
	}

	RedisConfig struct {
		Address  string `yaml:"address" env-required:"true"`
		Password string `yaml:"password" env-required:"true" env:"REDIS_PASSWORD"`
		DB       int    `yaml:"db" env-default:"0"`
	}

	HTTPConfig struct {
		Address string `yaml:"address" env-required:"true"`
	}

	MusicInfoConfig struct {
		Address string `yaml:"address" env-required:"true"`
	}
)

func MustLoad() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found")
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	// check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
