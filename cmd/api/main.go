package main

import (
	"log/slog"

	"github.com/Estriper0/avito_intership/internal/config"
	"github.com/Estriper0/avito_intership/internal/logger"
)

const configPath = "configs/config.yaml"

func main() {
	config := config.New(configPath)
	logger := logger.GetLogger(config.App.Env)
	logger.Info("", slog.Any("config", config))
}
