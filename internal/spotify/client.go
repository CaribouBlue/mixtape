package spotify

import (
	"net/http"
	"os"

	"github.com/CaribouBlue/mixtape/internal/core"
	"github.com/CaribouBlue/mixtape/internal/utils"
	"github.com/google/uuid"
)

type Client struct {
	accessToken  AccessToken
	ClientId     string
	ClientSecret string
	RedirectUri  string
	Scope        string
	currentUser  *UserProfile
}

func NewClient(clientId string, clientSecret string, redirectUri string, scope string) *Client {
	return &Client{
		ClientId:     clientId,
		ClientSecret: clientSecret,
		RedirectUri:  redirectUri,
		Scope:        scope,
	}
}

func NewDefaultClient() *Client {
	clientId := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	redirectUri := os.Getenv("SPOTIFY_REDIRECT_URI")
	scope := os.Getenv("SPOTIFY_SCOPE")

	spotifyClient := NewClient(clientId, clientSecret, redirectUri, scope)

	return spotifyClient
}

func (s *Client) GetUserAuthUrl() (string, error) {
	return GetUserAuthUrl(UserAuthRequestOptions{
		ClientId:    s.ClientId,
		RedirectUri: s.RedirectUri,
		Scope:       s.Scope,
		State:       uuid.New().String(),
	})
}

func (s *Client) Authenticate(code string) (AccessToken, error) {
	newAccessToken, err := GetAccessToken(AccessTokenRequestOptions{
		Code:         code,
		ClientId:     s.ClientId,
		ClientSecret: s.ClientSecret,
		RedirectUri:  s.RedirectUri,
		GrantType:    AuthorizationCodeGrantType,
	})
	if err != nil {
		return AccessToken{}, err
	}
	s.accessToken = *newAccessToken
	return s.GetValidAccessToken()
}

func (s *Client) refreshAccessToken() error {
	newAccessToken, err := GetAccessToken(AccessTokenRequestOptions{
		RefreshToken: s.accessToken.RefreshToken,
		ClientId:     s.ClientId,
		ClientSecret: s.ClientSecret,
		GrantType:    RefreshTokenGrantType,
	})
	if err != nil {
		return err
	}
	s.accessToken = *newAccessToken
	return nil
}

func (s *Client) Reauthenticate(refreshToken string) (AccessToken, error) {
	s.accessToken.RefreshToken = refreshToken
	err := s.refreshAccessToken()
	if err != nil {
		return AccessToken{}, err
	}

	return s.GetValidAccessToken()
}

func (s *Client) GetValidAccessToken() (AccessToken, error) {
	if s.accessToken.IsExpired() {
		err := s.refreshAccessToken()
		if err != nil {
			return s.accessToken, err
		}
	}

	return s.accessToken, nil
}

func (s *Client) AuthenticateUser(user *core.UserEntity) error {
	_, err := s.Reauthenticate(user.SpotifyToken)
	if err != nil {
		return err
	}
	return nil
}

func (s *Client) NewRequest(opts SpotifyRequestOptions) (*http.Request, error) {
	accessToken, err := s.GetValidAccessToken()
	if err != nil {
		return nil, err
	}

	return newRequest(SpotifyRequestOptions{
		accessToken: accessToken,
	})
}

func (s *Client) GetCurrentUserProfile() (*UserProfile, error) {
	accessToken, err := s.GetValidAccessToken()
	if err != nil {
		return nil, err
	}

	return getCurrentUserProfile(GetCurrentUserProfileRequestOptions{
		accessToken: accessToken,
	})
}

func (s *Client) CurrentUser() *UserProfile {
	if s.currentUser == nil {
		userProfile, err := s.GetCurrentUserProfile()
		if err == nil {
			s.currentUser = userProfile
		}
	}
	return s.currentUser
}

func (s *Client) SearchTracks(query string) ([]core.TrackEntity, error) {
	accessToken, err := s.GetValidAccessToken()
	if err != nil {
		return nil, err
	}

	results, err := getSearchResult(GetSearchResultRequestOptions{
		accessToken: accessToken,
		query:       query,
		itemTypes:   []ItemType{TrackItemType},
	})
	if err != nil {
		return nil, err
	}

	tracks := make([]core.TrackEntity, len(results.Tracks.Items))
	for i, track := range results.Tracks.Items {
		tracks[i] = core.TrackEntity{
			Id:   track.Id,
			Name: track.Name,
			Artists: utils.Map(track.Artists, func(artist SearchResultArtist) core.ArtistEntity {
				return core.ArtistEntity{Id: artist.Id, Name: artist.Name, Url: artist.ExternalUrls.Spotify}
			}),
			Album:    core.AlbumEntity{Id: track.Album.Id, Name: track.Album.Name, Url: track.Album.ExternalUrls.Spotify},
			Explicit: track.Explicit,
			Url:      track.ExternalUrls.Spotify,
		}
	}
	return tracks, nil
}

func (s *Client) GetTrackById(id string) (*core.TrackEntity, error) {
	accessToken, err := s.GetValidAccessToken()
	if err != nil {
		return nil, err
	}

	track, err := getTrack(GetTrackRequestOptions{
		accessToken: accessToken,
		id:          id,
	})
	if err != nil {
		return nil, err
	}

	return &core.TrackEntity{
		Id:   track.Id,
		Name: track.Name,
		Artists: utils.Map(track.Artists, func(artist TrackArtist) core.ArtistEntity {
			return core.ArtistEntity{Id: artist.Id, Name: artist.Name, Url: artist.ExternalUrls.Spotify}
		}),
		Album:    core.AlbumEntity{Id: track.Album.Id, Name: track.Album.Name, Url: track.Album.ExternalUrls.Spotify},
		Explicit: track.Explicit,
		Url:      track.ExternalUrls.Spotify,
	}, nil
}

func (s *Client) CreatePlaylist(name string, trackIds []string) (*core.PlaylistEntity, error) {
	accessToken, err := s.GetValidAccessToken()
	if err != nil {
		return nil, err
	}

	playlist, err := createPlaylist(CreatePlaylistRequestOptions{
		accessToken: accessToken,
		userId:      s.CurrentUser().Id,
		name:        name,
	})
	if err != nil {
		return nil, err
	}

	playlistEntity := &core.PlaylistEntity{
		Id:   playlist.Id,
		Name: playlist.Name,
		Url:  playlist.ExternalUrls.Spotify,
	}

	uris := utils.Map(trackIds, func(id string) string {
		return "spotify:track:" + id
	})

	err = addItemsToPlaylist(AddItemsToPlaylistRequestOptions{
		accessToken: accessToken,
		playlistId:  playlistEntity.Id,
		uris:        uris,
	})
	if err != nil {
		rollbackErr := unfollowPlaylist(UnfollowPlaylistRequestOptions{
			accessToken: accessToken,
			playlistId:  playlistEntity.Id,
		})
		if rollbackErr != nil {
			return playlistEntity, err
		} else {
			return nil, err
		}
	}

	return playlistEntity, nil
}

func (s *Client) GetPlaylistById(playlistId string) (*core.PlaylistEntity, error) {
	accessToken, err := s.GetValidAccessToken()
	if err != nil {
		return nil, err
	}

	playlist, err := getPlaylist(GetPlaylistRequestOptions{
		accessToken: accessToken,
		playlistId:  playlistId,
	})
	if err != nil {
		return nil, err
	}

	return &core.PlaylistEntity{
		Id:   playlist.Id,
		Name: playlist.Name,
		Url:  playlist.ExternalUrls.Spotify,
	}, nil
}
