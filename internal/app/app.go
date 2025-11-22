package app

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Estriper0/avito_intership/internal/config"
	"github.com/Estriper0/avito_intership/internal/handlers"
	"github.com/Estriper0/avito_intership/internal/repository/db"
	"github.com/Estriper0/avito_intership/internal/server"
	"github.com/Estriper0/avito_intership/internal/service"
	"github.com/Estriper0/avito_intership/pkg/postgres"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	logger *slog.Logger
	config *config.Config
	db     *pgxpool.Pool
	server *server.Server
}

func New(logger *slog.Logger, config *config.Config) *App {
	//Removing Gin logs in production
	if config.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	dbPool, err := postgres.New(config.DB.Url, config.DB.PoolSize)
	if err != nil {
		panic(err)
	}

	trManager := manager.Must(trmpgx.NewDefaultFactory(dbPool))
	validate := validator.New()

	teamRepo := db.NewTeamRepo(dbPool, trmpgx.DefaultCtxGetter)
	userRepo := db.NewUserRepo(dbPool, trmpgx.DefaultCtxGetter)
	prRepo := db.NewPullRequestRepo(dbPool, trmpgx.DefaultCtxGetter)

	teamService := service.NewTeamService(teamRepo, userRepo, trManager, logger)
	userService := service.NewUserService(userRepo, teamRepo, prRepo, logger)
	prService := service.NewPullRequestService(prRepo, userRepo, teamRepo, trManager, logger)

	teamGroup := router.Group("/team")
	handlers.NewTeamHandler(teamGroup, teamService, validate)

	userGroup := router.Group("/users")
	handlers.NewUserHandler(userGroup, userService, validate)

	prGroup := router.Group("pullRequest")
	handlers.NewPullRequestHandler(prGroup, prService, validate)

	server := server.New(router, config)

	return &App{
		logger: logger,
		config: config,
		db:     dbPool,
		server: server,
	}
}

func (a *App) Run() {
	//Closing the connection to the database
	defer a.db.Close()

	a.logger.Info("Start application")

	a.logger.Info(fmt.Sprintf("Starting server on :%d", a.config.Server.Port))
	go a.server.Run()

	quit := make(chan os.Signal, 1)
	defer close(quit)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	//Catch the application signal or error
	select {
	case q := <-quit:
		a.logger.Info(fmt.Sprintf("Received signal: %s", q.String()))
	case err := <-a.server.Err():
		a.logger.Error(fmt.Sprintf("Server error: %s", err.Error()))
	}
	a.logger.Info("Initiating graceful shutdown...")

	//Graceful shutdown
	err := a.server.Stop()
	if err != nil {
		a.logger.Error("Incorrect server shutdown", slog.String("error", err.Error()))
	} else {
		a.logger.Info("Server shutdown gracefully")
	}
	a.logger.Info("Stop application")
}
