package main

import (
	"log"
	"net/http"

	"github.com/yifen9/gamidoc-backend/config"
	"github.com/yifen9/gamidoc-backend/internal/app"
)

func main() {
	cfg := config.Load()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := application.Close(); err != nil {
			log.Print(err)
		}
	}()

	log.Printf("starting server on %s", cfg.HTTPAddr)

	if err := http.ListenAndServe(cfg.HTTPAddr, application.Router()); err != nil {
		log.Fatal(err)
	}
}
