package app

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Estriper0/avito_intership/internal/config"
	"github.com/Estriper0/avito_intership/internal/server"
	"github.com/gin-gonic/gin"
)

type App struct {
	logger *slog.Logger
	config *config.Config
	server *server.Server
}

func New(logger *slog.Logger, config *config.Config) *App {
	if config.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	server := server.New(router, config)

	return &App{
		logger: logger,
		config: config,
		server: server,
	}
}

func (a *App) Run() {
	a.logger.Info("Start application")

	a.logger.Info(fmt.Sprintf("Starting server on :%d", a.config.Server.Port))
	go a.server.Run()

	quit := make(chan os.Signal, 1)
	defer close(quit)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case q := <-quit:
		a.logger.Info(fmt.Sprintf("Received signal: %s", q.String()))
	case err := <-a.server.Err():
		a.logger.Error(fmt.Sprintf("Server error: %s", err.Error()))
	}
	a.logger.Info("Initiating graceful shutdown...")

	err := a.server.Stop()
	if err != nil {
		a.logger.Error("Incorrect server shutdown", slog.String("error", err.Error()))
	} else {
		a.logger.Info("Server shutdown gracefully")
	}
	a.logger.Info("Stop application")
}
