package main

import (
	"log"
	"net/http"

	"github.com/yifen9/gamidoc-backend/config"
	"github.com/yifen9/gamidoc-backend/internal/app"
)

func main() {
	cfg := config.Load()
	application := app.New(cfg)

	log.Printf("starting server on %s", cfg.HTTPAddr)

	if err := http.ListenAndServe(cfg.HTTPAddr, application.Router()); err != nil {
		log.Fatal(err)
	}
}
