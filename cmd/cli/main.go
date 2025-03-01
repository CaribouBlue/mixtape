package main

import (
	"log"

	"github.com/CaribouBlue/mixtape/cmd/cli/cmd"
	"github.com/CaribouBlue/mixtape/cmd/cli/config"
)

func init() {
	err := config.Load()
	if err != nil {
		log.Default().Println("No config loaded from env")
	}
}

func main() {
	cmd.Execute()
}
