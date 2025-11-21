package main

import (
	"log"
	"net/http"
	"os"

	"github.com/vninomtz/card-game/internal"
)

func main() {
	cfg := internal.NewConfig()
	err := cfg.Load()
	if err != nil {
		log.Fatalf("Error to load config: %v", err)
		os.Exit(1)
	}

	srv := internal.NewServer(cfg)

	if err := srv.Run(); err != nil && err != http.ErrServerClosed {
		log.Printf("Could not start the server: %v\n", err)
	}
}
