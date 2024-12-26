package spotify

import (
	"net/http"
	"os"

	"github.com/CaribouBlue/top-spot/internal/utils"
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

func DefaultClient() *Client {
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

func (s *Client) GetNewAccessToken(code string) error {
	newAccessToken, err := GetAccessToken(AccessTokenRequestOptions{
		Code:         code,
		ClientId:     s.ClientId,
		ClientSecret: s.ClientSecret,
		RedirectUri:  s.RedirectUri,
		GrantType:    AuthorizationCodeGrantType,
	})
	if err != nil {
		return err
	}
	s.accessToken = *newAccessToken
	return nil
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

func (s *Client) GetValidAccessToken() (AccessToken, error) {
	if s.accessToken.IsExpired() {
		err := s.refreshAccessToken()
		if err != nil {
			return s.accessToken, err
		}
	}

	return s.accessToken, nil
}

func (s *Client) SetAccessToken(accessToken AccessToken) {
	s.accessToken = accessToken
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

func (s *Client) SearchTracks(query string) (*SearchResult, error) {
	accessToken, err := s.GetValidAccessToken()
	if err != nil {
		return nil, err
	}

	return getSearchResult(GetSearchResultRequestOptions{
		accessToken: accessToken,
		query:       query,
		itemTypes:   []ItemType{TrackItemType},
	})
}

func (s *Client) GetTrack(id string) (*Track, error) {
	accessToken, err := s.GetValidAccessToken()
	if err != nil {
		return nil, err
	}

	return getTrack(GetTrackRequestOptions{
		accessToken: accessToken,
		id:          id,
	})
}

func (s *Client) CreatePlaylist(name string) (*Playlist, error) {
	accessToken, err := s.GetValidAccessToken()
	if err != nil {
		return nil, err
	}

	return createPlaylist(CreatePlaylistRequestOptions{
		accessToken: accessToken,
		userId:      s.CurrentUser().Id,
		name:        name,
	})
}

func (s *Client) AddTracksToPlaylist(playlistId string, trackIds []string) error {
	accessToken, err := s.GetValidAccessToken()
	if err != nil {
		return err
	}

	uris := utils.Map(trackIds, func(id string) string {
		return "spotify:track:" + id
	})

	return addItemsToPlaylist(AddItemsToPlaylistRequestOptions{
		accessToken: accessToken,
		playlistId:  playlistId,
		uris:        uris,
	})
}

func (s *Client) GetPlaylist(playlistId string) (*Playlist, error) {
	accessToken, err := s.GetValidAccessToken()
	if err != nil {
		return nil, err
	}

	return getPlaylist(GetPlaylistRequestOptions{
		accessToken: accessToken,
		playlistId:  playlistId,
	})
}

func (s *Client) UnfollowPlaylist(playlistId string) error {
	accessToken, err := s.GetValidAccessToken()
	if err != nil {
		return err
	}

	return unfollowPlaylist(UnfollowPlaylistRequestOptions{
		accessToken: accessToken,
		playlistId:  playlistId,
	})
}
