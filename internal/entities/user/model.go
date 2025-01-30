package user

import "github.com/CaribouBlue/top-spot/internal/spotify"

type User struct {
	Id                 int64               `json:"id"`
	UserName           string              `json:"userName"`
	PasswordHash       []byte              `json:"passwordHash"`
	SpotifyAccessToken spotify.AccessToken `json:"spotifyAccessToken"`
}
