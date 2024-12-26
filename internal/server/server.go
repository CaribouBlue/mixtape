package server

import (
	"fmt"
	"net/http"
	"os"

	"github.com/CaribouBlue/top-spot/internal/server/middleware"
	"github.com/CaribouBlue/top-spot/internal/server/mux"
)

func StartServer() {
	rootMux := mux.NewRootMux()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	serverAddr := fmt.Sprintf("localhost:%s", port)

	server := &http.Server{
		Addr:    serverAddr,
		Handler: middleware.Apply(rootMux, middleware.WithRequestLogging()),
	}

	fmt.Println("Starting server at", serverAddr)
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
