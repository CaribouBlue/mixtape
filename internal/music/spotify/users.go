package spotify

import (
	"encoding/json"
	"net/http"
)

type GetCurrentUserProfileRequestOptions struct {
	accessToken AccessToken
}

type UserProfile struct {
	Country         string `json:"country"`
	DisplayName     string `json:"display_name"`
	Email           string `json:"email"`
	ExplicitContent struct {
		FilterEnabled bool `json:"filter_enabled"`
		FilterLocked  bool `json:"filter_locked"`
	} `json:"explicit_content"`
	ExternalUrls struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Followers struct {
		Href  string `json:"href"`
		Total int    `json:"total"`
	} `json:"followers"`
	Href   string `json:"href"`
	Id     string `json:"id"`
	Images []struct {
		Url    string `json:"url"`
		Height int    `json:"height"`
		Width  int    `json:"width"`
	} `json:"images"`
	Product string `json:"product"`
	Type    string `json:"type"`
	Uri     string `json:"uri"`
}

func getCurrentUserProfile(opts GetCurrentUserProfileRequestOptions) (*UserProfile, error) {
	var profile UserProfile

	req, err := newRequest(SpotifyRequestOptions{
		method:      "GET",
		path:        "/me",
		accessToken: opts.accessToken,
	})

	if err != nil {
		return &profile, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return &profile, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&profile)
	if err != nil {
		return &profile, err
	}

	return &profile, nil
}
