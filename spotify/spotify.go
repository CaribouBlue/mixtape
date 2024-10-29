package spotify

import (
	"net/http"

	"github.com/CaribouBlue/top-spot/utils"
	"github.com/google/uuid"
)

type SpotifyClient struct {
	accessToken  AccessToken
	ClientId     string
	ClientSecret string
	RedirectUri  string
	Scope        string
	currentUser  *UserProfile
}

func (s *SpotifyClient) GetUserAuthUrl() (string, error) {
	return GetUserAuthUrl(UserAuthRequestOptions{
		ClientId:    s.ClientId,
		RedirectUri: s.RedirectUri,
		Scope:       s.Scope,
		State:       uuid.New().String(),
	})
}

func (s *SpotifyClient) GetNewAccessToken(code string) error {
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

func (s *SpotifyClient) refreshAccessToken() error {
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

func (s *SpotifyClient) GetValidAccessToken() (AccessToken, error) {
	if s.accessToken.IsExpired() {
		err := s.refreshAccessToken()
		if err != nil {
			return s.accessToken, err
		}
	}

	return s.accessToken, nil
}

func (s *SpotifyClient) SetAccessToken(accessToken AccessToken) {
	s.accessToken = accessToken
}

func (s *SpotifyClient) NewRequest(opts SpotifyRequestOptions) (*http.Request, error) {
	accessToken, err := s.GetValidAccessToken()
	if err != nil {
		return nil, err
	}

	return newRequest(SpotifyRequestOptions{
		accessToken: accessToken,
	})
}

func (s *SpotifyClient) GetCurrentUserProfile() (*UserProfile, error) {
	accessToken, err := s.GetValidAccessToken()
	if err != nil {
		return nil, err
	}

	return getCurrentUserProfile(GetCurrentUserProfileRequestOptions{
		accessToken: accessToken,
	})
}

func (s *SpotifyClient) CurrentUser() *UserProfile {
	if s.currentUser == nil {
		userProfile, err := s.GetCurrentUserProfile()
		if err == nil {
			s.currentUser = userProfile
		}
	}
	return s.currentUser
}

func (s *SpotifyClient) SearchTracks(query string) (*SearchResult, error) {
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

func (s *SpotifyClient) GetTrack(id string) (*Track, error) {
	accessToken, err := s.GetValidAccessToken()
	if err != nil {
		return nil, err
	}

	return getTrack(GetTrackRequestOptions{
		accessToken: accessToken,
		id:          id,
	})
}

func (s *SpotifyClient) CreatePlaylist(name string) (*Playlist, error) {
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

func (s *SpotifyClient) AddTracksToPlaylist(playlistId string, trackIds []string) error {
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

func NewSpotifyClient(clientId string, clientSecret string, redirectUri string, scope string) *SpotifyClient {
	return &SpotifyClient{
		ClientId:     clientId,
		ClientSecret: clientSecret,
		RedirectUri:  redirectUri,
		Scope:        scope,
	}
}
