package app

import (
	"context"
	"net/http"
	"time"

	"github.com/yifen9/gamidoc-backend/config"
	apphttp "github.com/yifen9/gamidoc-backend/internal/http"
	"github.com/yifen9/gamidoc-backend/internal/storage/postgres"
	rediscache "github.com/yifen9/gamidoc-backend/internal/storage/redis"
)

type App struct {
	config config.Config
	router http.Handler
	pg     *postgres.DB
	redis  *rediscache.Client
}

func New(cfg config.Config) (*App, error) {
	pg, err := postgres.New(cfg.PostgresDSN())
	if err != nil {
		return nil, err
	}

	redisClient := rediscache.New(cfg.RedisAddr())

	application := &App{
		config: cfg,
		pg:     pg,
		redis:  redisClient,
	}

	application.router = apphttp.NewRouter(apphttp.Dependencies{
		Postgres: application.pg,
		Redis:    application.redis,
	})

	return application, nil
}

func (a *App) Router() http.Handler {
	return a.router
}

func (a *App) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = a.pg.Ping(ctx)
	_ = a.redis.Ping(ctx)

	if err := a.pg.Close(); err != nil {
		return err
	}

	if err := a.redis.Close(); err != nil {
		return err
	}

	return nil
}
