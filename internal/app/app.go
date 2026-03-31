package app

import (
	"net/http"

	"github.com/yifen9/gamidoc-backend/config"
	apphttp "github.com/yifen9/gamidoc-backend/internal/http"
)

type App struct {
	config config.Config
	router http.Handler
}

func New(cfg config.Config) App {
	return App{
		config: cfg,
		router: apphttp.NewRouter(),
	}
}

func (a App) Router() http.Handler {
	return a.router
}
