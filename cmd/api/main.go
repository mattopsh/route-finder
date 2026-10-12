package main

import (
	"fmt"
	"log"

	"route-finder/internal/app"
	"route-finder/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	engine, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := engine.Run(fmt.Sprintf(":%d", cfg.ListenAddr)); err != nil {
		log.Fatal(err)
	}
}
