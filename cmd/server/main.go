package main

import (
	"log"

	"github.com/CaribouBlue/mixtape/internal/config"
	"github.com/CaribouBlue/mixtape/internal/server"
)

func main() {
	config.Load()
	s := server.NewServer()

	log.Default().Printf("Starting server at %s", s.Addr)
	if err := s.ListenAndServe(); err != nil {
		log.Fatalln("Error starting server:", err)
	}
}
