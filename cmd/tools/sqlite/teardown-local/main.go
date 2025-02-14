package main

import (
	"log"
	"os"
	"path"

	"github.com/CaribouBlue/top-spot/internal/appdata"
)

func main() {
	appDataDir, err := appdata.GetAppDataDir()
	if err != nil {
		log.Fatal(err)
	}

	dbPath := path.Join(appDataDir, "/top-spot.db")
	e := os.Remove(dbPath)
	if e != nil && !os.IsNotExist(e) {
		log.Fatal(e)
	}
}
