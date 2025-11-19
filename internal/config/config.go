package config

import (
	"fmt"
	"path"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	App AppConfig `yaml:"app"`
	DB  DBConfig  `yaml:"db"`
}

type AppConfig struct {
	Env  string `env:"ENV" env-default:"local"`
	Port int    `env-required:"true" yaml:"port" env:"APP_PORT"`
}

type DBConfig struct {
	Host     string `env-required:"true" env:"DB_HOST"`
	Port     string `env-required:"true" env:"DB_PORT"`
	User     string `env-required:"true" env:"DB_USER"`
	Password string `env-required:"true" env:"DB_PASSWORD"`
	Name     string `env-required:"true" env:"DB_NAME"`
	PoolSize int    `env-required:"true" yaml:"pool_size" env:"DB_POOL_SIZE"`
}

func New(configPath string) *Config {
	var config Config

	_ = godotenv.Load()

	err := cleanenv.ReadConfig(path.Join("./", configPath), &config)
	if err != nil {
		panic(fmt.Sprintf("error reading config: %s", err.Error()))
	}

	return &config
}
