package server

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/CaribouBlue/top-spot/internal/appdata"
	"github.com/CaribouBlue/top-spot/internal/db"
	"github.com/CaribouBlue/top-spot/internal/music"
	"github.com/CaribouBlue/top-spot/internal/server/middleware"
	"github.com/CaribouBlue/top-spot/internal/server/mux"
	"github.com/CaribouBlue/top-spot/internal/session"
	"github.com/CaribouBlue/top-spot/internal/user"
)

func StartServer() {
	// Initialize DB
	appDataDir, err := appdata.GetAppDataDir()
	if err != nil {
		log.Fatal("Error getting app data directory:", err)
	}

	dbPath := appDataDir + "/top-spot.db"
	sqliteJsonDb, err := db.NewSqliteJsonDb(dbPath)
	if err != nil {
		log.Fatal("Error creating SQLite JSON DB:", err)
	}

	// Initialize services
	userService := user.NewUserService(sqliteJsonDb)
	musicService := music.NewSpotifyMusicService()
	sessionService := session.NewSessionService(sqliteJsonDb, musicService)

	// Initialize server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	serverAddr := fmt.Sprintf("localhost:%s", port)

	rootMuxHandler := middleware.Apply(mux.NewRootMux(userService, musicService, sessionService), middleware.WithRequestLogging())

	server := &http.Server{
		Addr:    serverAddr,
		Handler: rootMuxHandler,
	}

	fmt.Println("Starting server at", serverAddr)
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
