package main

import (
	"log"

	"github.com/CaribouBlue/mixtape/cmd/cli/cmd"
	"github.com/CaribouBlue/mixtape/cmd/cli/config"
)

func init() {
	err := config.Load()
	if err != nil {
		log.Default().Fatalln("Unable to load default .env file: ", err)
	}
}

func main() {
	cmd.Execute()
}
