package main

import (
	"net/http"

	"github.com/yifen9/gamidoc-backend/config"
	"github.com/yifen9/gamidoc-backend/internal/app"
)

func main() {
	cfg := config.Load()

	application, err := app.New(cfg)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := application.Close(); err != nil {
			application.Logger().Error("app_close_failed", "error", err.Error())
		}
	}()

	application.Logger().Info("server_starting", "http_addr", cfg.HTTPAddr, "app_env", cfg.AppEnv)

	if err := http.ListenAndServe(cfg.HTTPAddr, application.Router()); err != nil {
		application.Logger().Error("server_stopped", "error", err.Error())
		panic(err)
	}
}
