package spotify

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	netUrl "net/url"
)

type GetTrackRequestOptions struct {
	accessToken AccessToken
	id          string
	market      string
}

type TrackArtist struct {
	ExternalUrls struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Href string `json:"href"`
	Id   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Uri  string `json:"uri"`
}

type Track struct {
	Album struct {
		AlbumType        string   `json:"album_type"`
		TotalTracks      int      `json:"total_tracks"`
		AvailableMarkets []string `json:"available_markets"`
		ExternalUrls     struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href   string `json:"href"`
		Id     string `json:"id"`
		Images []struct {
			Url    string `json:"url"`
			Height int    `json:"height"`
			Width  int    `json:"width"`
		} `json:"images"`
		Name                 string `json:"name"`
		ReleaseDate          string `json:"release_date"`
		ReleaseDatePrecision string `json:"release_date_precision"`
		Restrictions         struct {
			Reason string `json:"reason"`
		} `json:"restrictions"`
		Type    string `json:"type"`
		Uri     string `json:"uri"`
		Artists []struct {
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href string `json:"href"`
			Id   string `json:"id"`
			Name string `json:"name"`
			Type string `json:"type"`
			Uri  string `json:"uri"`
		} `json:"artists"`
	} `json:"album"`
	Artists          []TrackArtist `json:"artists"`
	AvailableMarkets []string      `json:"available_markets"`
	DiscNumber       int           `json:"disc_number"`
	DurationMs       int           `json:"duration_ms"`
	Explicit         bool          `json:"explicit"`
	ExternalIds      struct {
		Isrc string `json:"isrc"`
		Ean  string `json:"ean"`
		Upc  string `json:"upc"`
	} `json:"external_ids"`
	ExternalUrls struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Href         string   `json:"href"`
	Id           string   `json:"id"`
	IsPlayable   bool     `json:"is_playable"`
	LinkedFrom   struct{} `json:"linked_from"`
	Restrictions struct {
		Reason string `json:"reason"`
	} `json:"restrictions"`
	Name        string `json:"name"`
	Popularity  int    `json:"popularity"`
	PreviewUrl  string `json:"preview_url"`
	TrackNumber int    `json:"track_number"`
	Type        string `json:"type"`
	Uri         string `json:"uri"`
	IsLocal     bool   `json:"is_local"`
}

func getTrack(opts GetTrackRequestOptions) (*Track, error) {
	var track Track

	if opts.id == "" {
		return nil, errors.New("id is required")
	}

	params := netUrl.Values{}

	if opts.market != "" {
		params.Add("market", opts.market)
	}

	req, err := newRequest(SpotifyRequestOptions{
		method:      "GET",
		path:        fmt.Sprintf("/tracks/%s?", opts.id) + params.Encode(),
		accessToken: opts.accessToken,
	})
	if err != nil {
		return &track, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return &track, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		var body string
		if err != nil {
			body = ""
		}
		body = string(bodyBytes)
		return &track, fmt.Errorf("failed to get Spotify track with %s: %s", resp.Status, body)
	}

	err = json.NewDecoder(resp.Body).Decode(&track)
	if err != nil {
		return &track, err
	}

	return &track, nil
}
