package app

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/yifen9/gamidoc-backend/config"
	"github.com/yifen9/gamidoc-backend/internal/auth"
	apphttp "github.com/yifen9/gamidoc-backend/internal/http"
	appmiddleware "github.com/yifen9/gamidoc-backend/internal/http/middleware"
	"github.com/yifen9/gamidoc-backend/internal/project"
	"github.com/yifen9/gamidoc-backend/internal/session"
	"github.com/yifen9/gamidoc-backend/internal/storage/postgres"
	rediscache "github.com/yifen9/gamidoc-backend/internal/storage/redis"
	"github.com/yifen9/gamidoc-backend/internal/token"
)

type App struct {
	config config.Config
	logger *slog.Logger
	router http.Handler
	pg     *postgres.DB
	redis  *rediscache.Client
}

func New(cfg config.Config) (*App, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	pg, err := postgres.New(cfg.PostgresDSN())
	if err != nil {
		return nil, err
	}

	redisClient := rediscache.New(cfg.RedisAddr())
	tokenManager := token.NewManager(cfg.JWTSecret, cfg.JWTExpiresIn)
	appmiddleware.SetTokenManager(tokenManager)

	userRepository := postgres.NewUserRepository(pg)
	authService := auth.NewService(userRepository, tokenManager)
	authHandler := auth.NewHandler(authService)

	projectRepository := postgres.NewProjectRepository(pg)
	projectService := project.NewService(projectRepository)
	projectHandler := project.NewHandler(projectService)

	sessionRepository := rediscache.NewSessionRepository(redisClient, cfg.SessionTTL)
	sessionService := session.NewService(sessionRepository, cfg.SessionTTL)
	sessionHandler := session.NewHandler(sessionService)

	application := &App{
		config: cfg,
		logger: logger,
		pg:     pg,
		redis:  redisClient,
	}

	application.router = apphttp.NewRouter(apphttp.Dependencies{
		Logger:         application.logger,
		Postgres:       application.pg,
		Redis:          application.redis,
		AuthHandler:    authHandler,
		ProjectHandler: projectHandler,
		SessionHandler: sessionHandler,
	})

	return application, nil
}

func (a *App) Router() http.Handler {
	return a.router
}

func (a *App) Logger() *slog.Logger {
	return a.logger
}

func (a *App) Close() error {
	if err := a.pg.Close(); err != nil {
		return err
	}

	if err := a.redis.Close(); err != nil {
		return err
	}

	return nil
}
