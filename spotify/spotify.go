package spotify

import (
	"net/http"

	"github.com/google/uuid"
)

type SpotifyClient struct {
	accessToken  AccessToken
	ClientId     string
	ClientSecret string
	RedirectUri  string
	Scope        string
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

func NewSpotifyClient(clientId string, clientSecret string, redirectUri string, scope string) *SpotifyClient {
	return &SpotifyClient{
		ClientId:     clientId,
		ClientSecret: clientSecret,
		RedirectUri:  redirectUri,
		Scope:        scope,
	}
}
