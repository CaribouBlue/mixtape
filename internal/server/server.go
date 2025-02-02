package server

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/CaribouBlue/top-spot/internal/appdata"
	"github.com/CaribouBlue/top-spot/internal/db"
	"github.com/CaribouBlue/top-spot/internal/entities/music"
	"github.com/CaribouBlue/top-spot/internal/entities/session"
	"github.com/CaribouBlue/top-spot/internal/entities/user"
	"github.com/CaribouBlue/top-spot/internal/server/middleware"
	"github.com/CaribouBlue/top-spot/internal/server/mux"
)

func StartServer() {
	// Initialize DB
	appDataDir, err := appdata.GetAppDataDir()
	if err != nil {
		log.Fatal("Error getting app data directory:", err)
	}

	dbPath := appDataDir + "/top-spot.db"
	db, err := db.NewSqliteJsonDb(dbPath)
	if err != nil {
		log.Fatal("Error creating SQLite JSON DB:", err)
	}

	// Initialize services
	userService := user.NewUserService(db)
	musicService := music.NewSpotifyMusicService()
	sessionService := session.NewSessionService(db)

	// Initialize server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	serverAddr := fmt.Sprintf("localhost:%s", port)

	rootMuxHandler := mux.NewRootMux(
		mux.RootMuxServices{
			UserService: userService,
		},
		[]middleware.Middleware{
			middleware.WithRequestMetadata(),
			middleware.WithRequestLogging(),
			middleware.WithCustomNotFoundHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Not found", http.StatusNotFound)
			})),
		},
		mux.RootMuxChildren{
			AuthMux: mux.NewAuthMux(
				mux.AuthMuxOpts{
					PathPrefix:       "/auth",
					LoginSuccessPath: "/app",
				},
				mux.AuthMuxServices{
					UserService:  userService,
					MusicService: musicService,
				},
				[]middleware.Middleware{
					middleware.WithUser(middleware.WithUserOpts{
						DefaultUserId: 6666,
						UserService:   userService,
					}),
				},
				mux.AuthMuxChildren{},
			),
			AppMux: mux.NewAppMux(
				mux.AppMuxOpts{
					PathPrefix: "/app",
				},
				mux.AppMuxServices{},
				[]middleware.Middleware{
					middleware.WithUser(middleware.WithUserOpts{
						DefaultUserId: 6666,
						UserService:   userService,
					}),
					middleware.WithEnforcedAuthentication(middleware.WithEnforcedAuthenticationOpts{
						UnauthenticatedRedirectPath: "/auth/login",
						UserService:                 userService,
					}),
					middleware.WithCustomNotFoundHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						http.Redirect(w, r, "/app/session/", http.StatusFound)
					})),
				},
				mux.AppMuxChildren{
					SessionMux: mux.NewSessionMux(
						mux.SessionMuxOpts{
							PathPrefix: "/session",
						},
						mux.SessionMuxServices{
							SessionService: sessionService,
							MusicService:   musicService,
							UserService:    userService,
						},
						[]middleware.Middleware{},
						mux.SessionMuxChildren{},
					),
					ProfileMux: mux.NewProfileMux(
						mux.ProfileMuxOpts{
							PathPrefix: "/profile",
						},
						mux.ProfileMuxServices{},
						[]middleware.Middleware{},
						mux.ProfileMuxChildren{},
					),
				},
			),
		},
	)

	server := &http.Server{
		Addr:    serverAddr,
		Handler: rootMuxHandler,
	}

	fmt.Println("Starting server at", serverAddr)
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
