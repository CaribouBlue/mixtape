package spotify

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	netUrl "net/url"
	"strings"
	"time"
)

type UserAuthRequestOptions struct {
	ClientId    string
	RedirectUri string
	Scope       string
	State       string
}

func GetUserAuthUrl(opts UserAuthRequestOptions) (string, error) {
	url := "https://accounts.spotify.com/authorize?"
	params := netUrl.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", opts.ClientId)
	params.Add("scope", opts.Scope)
	params.Add("redirect_uri", opts.RedirectUri)
	params.Add("state", opts.State)

	return url + params.Encode(), nil
}

type GrantType string

const (
	AuthorizationCodeGrantType GrantType = "authorization_code"
	RefreshTokenGrantType      GrantType = "refresh_token"
)

type AccessTokenRequestOptions struct {
	Code         string
	ClientId     string
	ClientSecret string
	RedirectUri  string
	State        string
	GrantType    GrantType
	RefreshToken string
}

type AccessToken struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	Scope        string    `json:"scope"`
	ExpiresIn    int       `json:"expires_in"`
	CreatedAt    time.Time `json:"created_at"`
	RefreshToken string    `json:"refresh_token"`
}

func (token *AccessToken) IsExpired() bool {
	if token.CreatedAt.IsZero() {
		return true
	}

	return time.Now().After(token.CreatedAt.Add(time.Duration(token.ExpiresIn) * time.Second))
}

func GetAccessToken(opts AccessTokenRequestOptions) (*AccessToken, error) {
	response := &AccessToken{
		RefreshToken: opts.RefreshToken,
		CreatedAt:    time.Now(),
	}

	data := netUrl.Values{}
	data.Set("code", opts.Code)
	data.Set("redirect_uri", opts.RedirectUri)
	data.Set("grant_type", string(opts.GrantType))
	data.Set("refresh_token", opts.RefreshToken)

	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(opts.ClientId+":"+opts.ClientSecret))

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		return response, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", authHeader)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return response, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	if resp.StatusCode != 200 {
		return response, fmt.Errorf("%d %s", resp.StatusCode, body)
	}

	json.Unmarshal(body, response)

	return response, nil
}
