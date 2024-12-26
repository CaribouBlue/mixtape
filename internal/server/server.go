package server

import (
	"fmt"
	"net/http"
	"os"

	"github.com/CaribouBlue/top-spot/internal/server/middleware"
	"github.com/CaribouBlue/top-spot/internal/server/mux"
)

func StartServer() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	serverAddr := fmt.Sprintf("localhost:%s", port)

	rootMuxHandler := middleware.Apply(mux.NewRootMux(), middleware.WithRequestLogging())

	server := &http.Server{
		Addr:    serverAddr,
		Handler: rootMuxHandler,
	}

	fmt.Println("Starting server at", serverAddr)
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
