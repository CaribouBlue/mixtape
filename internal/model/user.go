package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/CaribouBlue/top-spot/internal/db"
	"github.com/CaribouBlue/top-spot/internal/spotify"
)

var (
	ErrUserPlaylistNotFound = errors.New("no playlist found with the given ID")
	ErrUserPlaylistExists   = errors.New("playlist already exists")
)

type UserPlaylistData struct {
	Id        string `json:"id"`
	SessionId int64  `json:"sessionId"`
}

type UserData struct {
	Id                 int64               `json:"id"`
	SpotifyAccessToken spotify.AccessToken `json:"spotifyAccessToken"`
	Playlists          []UserPlaylistData  `json:"playlists"`
}

type UserModel struct {
	Data UserData
	db   db.Db
}

func NewUserModel(db db.Db, opts ...OptsFn) *UserModel {
	user := &UserModel{
		db: db,
	}

	for _, opt := range opts {
		_ = opt(user)
	}

	return user
}

func (user *UserModel) Name() string {
	return "user"
}

func (user *UserModel) Id() int64 {
	return user.Data.Id
}

func (user *UserModel) SetId(id int64) {
	user.Data.Id = id
}

func (user *UserModel) Scan(value interface{}) error {
	return json.Unmarshal([]byte(value.(string)), &user.Data)
}

func (user *UserModel) Value() (driver.Value, error) {
	return json.Marshal(user.Data)
}

func (user *UserModel) Create() error {
	return user.db.CreateRecord(user)
}

func (user *UserModel) Read() error {
	return user.db.ReadRecord(user)
}

func (user *UserModel) Update() error {
	return user.db.UpdateRecord(user)
}

func (user *UserModel) AddPlaylist(playlistId string, sessionId int64) error {
	_, err := user.GetPlaylist(playlistId)
	if err == nil {
		return ErrUserPlaylistExists
	} else if err != ErrUserPlaylistNotFound {
		return err
	}

	user.Data.Playlists = append(user.Data.Playlists, UserPlaylistData{Id: playlistId, SessionId: sessionId})
	return nil
}

func (user *UserModel) GetPlaylist(playlistId string) (UserPlaylistData, error) {
	for _, playlist := range user.Data.Playlists {
		if playlist.Id == playlistId {
			return playlist, nil
		}
	}

	return UserPlaylistData{}, ErrUserPlaylistNotFound
}

func (user *UserModel) GetSessionPlaylist(sessionId int64) (UserPlaylistData, error) {
	for _, playlist := range user.Data.Playlists {
		if playlist.SessionId == sessionId {
			return playlist, nil
		}
	}

	return UserPlaylistData{}, ErrUserPlaylistNotFound
}

func (user *UserModel) IsAuthenticated() (bool, error) {
	if user.Data.SpotifyAccessToken.AccessToken == "" {
		return false, nil
	}

	for _, scope := range strings.Split(os.Getenv("SPOTIFY_SCOPE"), " ") {
		if !strings.Contains(user.Data.SpotifyAccessToken.Scope, scope) {
			return false, nil
		}
	}

	return true, nil
}
