package user

import (
	"strconv"

	"github.com/CaribouBlue/top-spot/internal/spotify"
)

type ByUsername []User

func (a ByUsername) Len() int           { return len(a) }
func (a ByUsername) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByUsername) Less(i, j int) bool { return a[i].Username < a[j].Username }

type User struct {
	Id                 int64               `json:"id"`
	Username           string              `json:"username"`
	PasswordHash       []byte              `json:"passwordHash"`
	SpotifyAccessToken spotify.AccessToken `json:"spotifyAccessToken"`
}

func (u *User) IdString() string {
	return strconv.FormatInt(u.Id, 10)
}
