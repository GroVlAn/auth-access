package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type HTTP struct {
	Port              string        `yaml:"port" env-default:"9080"`
	MaxHeaderBytes    int           `yaml:"max_header_bytes" env-default:"4096"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout" env-default:"10s"`
	WriteTimeout      time.Duration `yaml:"write_timeout" env-default:"10s"`
	BaseHTTPPath      string        `yaml:"base_http_path" env-default:"/api"`
}

type GRPC struct {
	Port string `yaml:"port"`
}

type PostgresSettings struct {
	Host     string `yaml:"host" env-required:"true"`
	Port     string `yaml:"port"`
	Username string `env:"DB_USERNAME" env-required:"true"`
	Password string `env:"DB_PASSWORD" env-required:"true"`
	DBName   string `env:"DB_NAME" env-required:"true"`
	SSLMode  string `yaml:"ssl_mode"`
}

type Settings struct {
	DefaultTimeout time.Duration `yaml:"default_timeout"`
}

type Config struct {
	HTTP     HTTP             `yaml:"http"`
	GRPC     GRPC             `yaml:"grpc"`
	Settings Settings         `yaml:"settings"`
	DB       PostgresSettings `yaml:"db"`
}

func New(path string) (*Config, error) {
	cfg := &Config{}

	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return nil, fmt.Errorf("reading config file %s: %w", path, err)
	}

	return cfg, nil
}

func LoadEnv(filenames ...string) error {
	if len(filenames) == 0 {
		return godotenv.Load()
	}

	for _, filename := range filenames {
		if err := godotenv.Load(filename); err != nil {
			return fmt.Errorf("loading env file %s: %w", filename, err)
		}
	}
	return nil
}
