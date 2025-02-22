package server

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/CaribouBlue/top-spot/internal/core"
	"github.com/CaribouBlue/top-spot/internal/mail"
	"github.com/CaribouBlue/top-spot/internal/server/middleware"
	"github.com/CaribouBlue/top-spot/internal/server/mux"
	"github.com/CaribouBlue/top-spot/internal/server/utils"
	"github.com/CaribouBlue/top-spot/internal/spotify"
	"github.com/CaribouBlue/top-spot/internal/storage"
)

func StartServer() {
	// Initialize DB
	dbPath := os.Getenv("DB_PATH")
	db, err := storage.NewSqliteDb(dbPath)
	if err != nil {
		log.Fatal("Error creating SQLite DB:", err)
	}

	// Initialize Mailer
	mailer, err := mail.NewGmailMailer(os.Getenv("GMAIL_USERNAME"), os.Getenv("GMAIL_PASSWORD"))
	if err != nil {
		log.Fatal("Error creating Gmail mailer:", err)
	}

	// Initialize services
	userService := core.NewUserService(db)
	_ = mail.NewMailService(mailer)

	// Initialize server
	serverAddress := os.Getenv("SERVER_ADDRESS")
	if serverAddress == "" {
		serverAddress = "localhost:8080"
	}

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
			StaticMux: mux.NewStaticMux(
				mux.StaticMuxOpts{
					PathPrefix: "/static",
				},
				[]middleware.Middleware{
					middleware.WithRequestMetadata(),
				},
			),
			AuthMux: mux.NewAuthMux(
				mux.AuthMuxOpts{
					PathPrefix:       "/auth",
					LoginSuccessPath: "/app/home",
				},
				mux.AuthMuxServices{
					UserService: userService,
				},
				[]middleware.Middleware{
					middleware.WithUser(middleware.WithUserOpts{
						DefaultUserId: 6666,
						UserService:   userService,
					}),
					middleware.WithSpotifyClient(),
					middleware.WithCustomNotFoundHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						utils.HandleRedirect(w, r, "/auth/login")
					})),
				},
			),
			AppMux: mux.NewAppMux(
				mux.AppMuxOpts{
					PathPrefix: "/app",
				},
				[]middleware.Middleware{
					middleware.WithUser(middleware.WithUserOpts{
						DefaultUserId: 6666,
						UserService:   userService,
					}),
					middleware.WithEnforcedAuthentication(middleware.WithEnforcedAuthenticationOpts{
						UnauthenticatedRedirectPath: "/auth/login",
						UserService:                 userService,
					}),
					middleware.WithSpotifyClient(),
					middleware.WithCustomNotFoundHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						utils.HandleRedirect(w, r, "/app/home")
					})),
				},
				mux.AppMuxChildren{
					SessionMux: mux.NewSessionMux(
						mux.SessionMuxOpts{
							PathPrefix: "/session",
						},
						mux.SessionMuxRepos{
							SessionRepo:      db,
							MusicRepoFactory: musicRepoFactory,
							UserRepo:         db,
						},
						[]middleware.Middleware{},
					),
					ProfileMux: mux.NewProfileMux(
						mux.ProfileMuxOpts{
							PathPrefix: "/profile",
						},
						[]middleware.Middleware{},
					),
				},
			),
		},
	)

	server := &http.Server{
		Addr:    serverAddress,
		Handler: rootMuxHandler,
	}

	fmt.Println("Starting server at", serverAddress)
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

var musicRepoFactory utils.RequestBasedFactory[core.MusicRepository] = func(r *http.Request) core.MusicRepository {
	user := r.Context().Value(utils.UserCtxKey).(*core.UserEntity)
	spotifyClient := spotify.NewDefaultClient()

	// TODO: handle invalid token
	if user.SpotifyToken != "" {
		spotifyClient.Reauthenticate(user.SpotifyToken)
	}

	return spotifyClient
}
