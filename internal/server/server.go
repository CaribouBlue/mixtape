package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/CaribouBlue/mixtape/internal/config"
	"github.com/CaribouBlue/mixtape/internal/core"
	"github.com/CaribouBlue/mixtape/internal/mail"
	"github.com/CaribouBlue/mixtape/internal/server/middleware"
	"github.com/CaribouBlue/mixtape/internal/server/mux"
	"github.com/CaribouBlue/mixtape/internal/server/utils"
	"github.com/CaribouBlue/mixtape/internal/spotify"
	"github.com/CaribouBlue/mixtape/internal/storage"
)

func NewServer() *http.Server {
	// Initialize DB
	dbPath := config.GetConfigValue(config.ConfDbPath)
	db, err := storage.NewSqliteDb(dbPath)
	if err != nil {
		log.Fatal("Error creating SQLite DB:", err)
	}

	// Initialize Mailer
	mailer, err := mail.NewGmailMailer(
		config.GetConfigValue(config.ConfGmailUsername),
		config.GetConfigValue(config.ConfGmailPassword),
	)
	if err != nil {
		log.Fatal("Error creating Gmail mailer:", err)
	}

	// Initialize services
	userService := core.NewUserService(db)
	_ = mail.NewMailService(mailer)

	// Initialize server
	host := config.GetConfigValue(config.ConfHost)
	port := config.GetConfigValue(config.ConfPort)
	serverAddress := fmt.Sprintf("%s:%s", host, port)
	if serverAddress == ":" {
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

	return server
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
