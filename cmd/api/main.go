package main

import (
	"github.com/Estriper0/avito_intership/internal/app"
	"github.com/Estriper0/avito_intership/internal/config"
	"github.com/Estriper0/avito_intership/internal/logger"
)

const configPath = "configs/config.yaml"

func main() {
	config := config.New(configPath)
	logger := logger.GetLogger(config.App.Env)
	app := app.New(logger, config)
	app.Run()
}
