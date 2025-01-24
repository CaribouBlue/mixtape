package user

import "github.com/CaribouBlue/top-spot/internal/spotify"

type User struct {
	Id                 int64               `json:"id"`
	SpotifyAccessToken spotify.AccessToken `json:"spotifyAccessToken"`
}
