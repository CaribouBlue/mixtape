package db

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/CaribouBlue/top-spot/spotify"
)

var (
	ErrUserPlaylistNotFound = errors.New("no playlist found with the given ID")
	ErrUserPlaylistExists   = errors.New("playlist already exists")
)

type UserPlaylist struct {
	Id        string `json:"id"`
	SessionId int64  `json:"sessionId"`
}

type UserDataModel struct {
	Id                 int64               `json:"id"`
	SpotifyAccessToken spotify.AccessToken `json:"spotifyAccessToken"`
	Playlists          []UserPlaylist      `json:"playlists"`
}

func (user *UserDataModel) GetTableName() string {
	return "user"
}

func (user *UserDataModel) SetId(id int64) {
	user.Id = id
}

func (user *UserDataModel) GetId() int64 {
	return user.Id
}

func (user *UserDataModel) Scan(value interface{}) error {
	return json.Unmarshal([]byte(value.(string)), user)
}

func (user *UserDataModel) Value() (driver.Value, error) {
	return json.Marshal(user)
}

func (user *UserDataModel) Insert() error {
	_, err := insertJsonDataModel(user)
	return err
}

func (user *UserDataModel) Update() error {
	return updateJsonDataModel(user)
}

func (user *UserDataModel) GetById() error {
	_, err := getJsonDataModelById(user)
	return err
}

func (user *UserDataModel) IsAuthenticated() (bool, error) {
	err := user.GetById()
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}

	if user.SpotifyAccessToken.AccessToken == "" {
		return false, nil
	}

	for _, scope := range strings.Split(os.Getenv("SPOTIFY_SCOPE"), " ") {
		if !strings.Contains(user.SpotifyAccessToken.Scope, scope) {
			return false, nil
		}
	}

	return true, nil
}

func (user *UserDataModel) GetPlaylist(playlistId string) (UserPlaylist, error) {
	for _, playlist := range user.Playlists {
		if playlist.Id == playlistId {
			return playlist, nil
		}
	}

	return UserPlaylist{}, ErrUserPlaylistNotFound
}

func (user *UserDataModel) GetPlaylistBySessionId(sessionId int64) (UserPlaylist, error) {
	for _, playlist := range user.Playlists {
		if playlist.SessionId == sessionId {
			return playlist, nil
		}
	}

	return UserPlaylist{}, ErrUserPlaylistNotFound
}

func (user *UserDataModel) AddPlaylist(playlistId string, sessionId int64) error {
	_, err := user.GetPlaylist(playlistId)
	if err == nil {
		return ErrUserPlaylistExists
	} else if err != ErrUserPlaylistNotFound {
		return err
	}

	user.Playlists = append(user.Playlists, UserPlaylist{Id: playlistId, SessionId: sessionId})
	return nil
}

func NewUserDataModel() *UserDataModel {
	user := UserDataModel{}
	return &user
}
