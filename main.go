package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	cfg := NewConfig()
	err := cfg.Load()
	if err != nil {
		log.Fatalf("Error to load config: %v", err)
		os.Exit(1)
	}

	srv := NewServer(cfg)

	if err := srv.Run(); err != nil && err != http.ErrServerClosed {
		log.Printf("Could not start the server: %v\n", err)
	}
}
