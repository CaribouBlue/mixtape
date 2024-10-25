package server

import (
	"fmt"
	"net/http"
	"os"
)

func StartServer() {
	rootMux := http.NewServeMux()

	// /auth
	registerAuthMux(rootMux)

	// /app
	registerAppMux(rootMux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	serverAddr := fmt.Sprintf("localhost:%s", port)
	fmt.Println("Starting server at ", serverAddr)
	if err := http.ListenAndServe(serverAddr, rootMux); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
