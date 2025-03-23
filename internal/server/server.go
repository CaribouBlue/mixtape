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
		mux.RootMuxOpts{},
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
		[]mux.ChildMux{
			mux.NewStaticMux(
				mux.StaticMuxOpts{
					MuxOpts: mux.MuxOpts{
						PathPrefix: "/static",
					},
				},
				mux.StaticMuxServices{},
				[]middleware.Middleware{},
				[]mux.ChildMux{},
			),
			mux.NewAuthMux(
				mux.AuthMuxOpts{
					MuxOpts: mux.MuxOpts{
						PathPrefix: "/auth",
					},
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
				[]mux.ChildMux{},
			),
			mux.NewAppMux(
				mux.AppMuxOpts{
					MuxOpts: mux.MuxOpts{
						PathPrefix: "/app",
					},
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
					middleware.WithSpotifyClient(),
					middleware.WithCustomNotFoundHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						utils.HandleRedirect(w, r, "/app/home")
					})),
				},
				[]mux.ChildMux{
					mux.NewSessionMux(
						mux.SessionMuxOpts{
							MuxOpts: mux.MuxOpts{
								PathPrefix: "/session",
							},
						},
						mux.SessionMuxServices{
							SessionServiceInitializer: func(mux *mux.SessionMux, r *http.Request) (*core.SessionService, error) {
								musicService, err := mux.Services.MusicService()
								if err != nil {
									return nil, err
								}

								return core.NewSessionService(db, mux.Services.UserService, musicService), nil
							},
							MusicServiceInitializer: func(mux *mux.SessionMux, r *http.Request) (*core.MusicService, error) {
								user, err := utils.ContextValue(r.Context(), utils.UserCtxKey)
								if err != nil {
									return nil, err
								}

								spotifyClient := spotify.NewDefaultClient()

								// TODO: handle invalid token
								if user.SpotifyToken != "" {
									spotifyClient.Reauthenticate(user.SpotifyToken)
								}

								return core.NewMusicService(spotifyClient), nil
							},
							UserService: userService,
						},
						[]middleware.Middleware{},
						[]mux.ChildMux{},
					),
					mux.NewProfileMux(
						mux.ProfileMuxOpts{
							MuxOpts: mux.MuxOpts{
								PathPrefix: "/profile",
							},
						},
						mux.ProfileMuxServices{},
						[]middleware.Middleware{},
						[]mux.ChildMux{},
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
