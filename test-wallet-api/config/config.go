package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	// Config -.
	Config struct {
		App   `yaml:"app"`
		HTTP  `yaml:"http"`
		Log   `yaml:"logger"`
		PG    `yaml:"postgres"`
		Redis `yaml:"redis"`
	}

	// App -.
	App struct {
		Name    string `env-required:"true" yaml:"name"    env:"APP_NAME"`
		Version string `env-required:"true" yaml:"version" env:"APP_VERSION"`
	}

	// HTTP -.
	HTTP struct {
		Port     string `env-required:"true" yaml:"port" env:"API_HTTP_PORT"`
		AuthPort string `env-required:"true" env:"AUTH_HTTP_PORT"`
		AuthHost string `env-required:"true" env:"AUTH_HTTP_HOST"`
	}

	// Log -.
	Log struct {
		Level string `env-required:"true" yaml:"log_level"   env:"LOG_LEVEL"`
	}

	Redis struct {
		Host     string `env-required:"true"                 env:"REDIS_HOST"`
		Username string `env-required:"true"                 env:"REDIS_USERNAME"`
		Password string `env-required:"true"                 env:"REDIS_PASSWORD"`
		Port     string `env-required:"true"                 env:"REDIS_PORT"`
	}
	// PG -.
	PG struct {
		Username string `env-required:"true"  yaml:"username"               env:"PG_USERNAME"`
		Password string `env-required:"true"                 env:"PG_PASSWORD"`
		Port     string `env-required:"true"                 env:"PG_PORT"`
		DbName   string `env-required:"true"   yaml:"dbName"               env:"PG_DBNANME"`
		HOST     string `env-required:"true"   yaml:"host"               env:"POSTGRES_HOST"`
	}
)

// NewConfig returns app config.
func NewConfig() (*Config, error) {
	cfg := &Config{}

	relativePath, _ := os.Getwd()
	fmt.Println()
	filepath.Join()
	err := cleanenv.ReadConfig(filepath.Join(relativePath, "../config", "config.yml"), cfg)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
