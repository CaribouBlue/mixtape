package server

import (
	"fmt"
	"net/http"
	"os"
)

func StartServer() {
	rootMux := http.NewServeMux()
	registerAuthMux(rootMux)
	registerAppMux(rootMux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	serverAddr := fmt.Sprintf("localhost:%s", port)

	server := &http.Server{
		Addr:    serverAddr,
		Handler: applyMiddleware(rootMux, withRequestLogging),
	}

	fmt.Println("Starting server at", serverAddr)
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
