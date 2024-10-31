package spotify

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type Playlist struct {
	Collaborative bool   `json:"collaborative"`
	Description   string `json:"description"`
	ExternalUrls  struct {
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
	Name  string `json:"name"`
	Owner struct {
		ExternalUrls struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Followers struct {
			Href  string `json:"href"`
			Total int    `json:"total"`
		} `json:"followers"`
		Href        string `json:"href"`
		Id          string `json:"id"`
		Type        string `json:"type"`
		Uri         string `json:"uri"`
		DisplayName string `json:"display_name"`
	} `json:"owner"`
	Public     bool   `json:"public"`
	SnapshotId string `json:"snapshot_id"`
	Tracks     struct {
		Href     string `json:"href"`
		Limit    int    `json:"limit"`
		Next     string `json:"next"`
		Offset   int    `json:"offset"`
		Previous string `json:"previous"`
		Total    int    `json:"total"`
		Items    []struct {
			AddedAt string `json:"added_at"`
			AddedBy struct {
				ExternalUrls struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
				Followers struct {
					Href  string `json:"href"`
					Total int    `json:"total"`
				} `json:"followers"`
				Href string `json:"href"`
				Id   string `json:"id"`
				Type string `json:"type"`
				Uri  string `json:"uri"`
			} `json:"added_by"`
			IsLocal bool `json:"is_local"`
			Track   struct {
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
				AvailableMarkets []string `json:"available_markets"`
				DiscNumber       int      `json:"disc_number"`
				DurationMs       int      `json:"duration_ms"`
				Explicit         bool     `json:"explicit"`
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
			} `json:"track"`
		} `json:"items"`
	} `json:"tracks"`
	Type string `json:"type"`
	Uri  string `json:"uri"`
}

type CreatePlaylistRequestOptions struct {
	accessToken   AccessToken
	userId        string
	name          string
	description   string
	public        bool
	collaborative bool
}

func createPlaylist(opts CreatePlaylistRequestOptions) (*Playlist, error) {
	var playlist Playlist

	if opts.userId == "" {
		return &playlist, errors.New("id is required")
	}

	if opts.name == "" {
		return &playlist, errors.New("name is required")
	}

	jsonData, err := json.Marshal(map[string]interface{}{
		"name":          opts.name,
		"description":   opts.description,
		"public":        opts.public,
		"collaborative": opts.collaborative,
	})
	if err != nil {
		return &playlist, err
	}
	body := bytes.NewReader(jsonData)

	req, err := newRequest(SpotifyRequestOptions{
		method:      "POST",
		path:        fmt.Sprintf("/users/%s/playlists", opts.userId),
		accessToken: opts.accessToken,
		body:        body,
	})
	if err != nil {
		return &playlist, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return &playlist, err
	}

	if resp.StatusCode != http.StatusCreated {
		return &playlist, fmt.Errorf("failed to create playlist")
	}

	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&playlist)
	if err != nil {
		return &playlist, err
	}

	return &playlist, nil
}

type AddItemsToPlaylistRequestOptions struct {
	accessToken AccessToken
	playlistId  string
	uris        []string
	position    int
}

func addItemsToPlaylist(opts AddItemsToPlaylistRequestOptions) error {
	if opts.playlistId == "" {
		return errors.New("playlist ID is required")
	}

	if len(opts.uris) < 1 {
		return errors.New("at least one URI is required")
	} else if len(opts.uris) > 100 {
		return errors.New("maximum of 100 URIs allowed")
	}

	jsonBody := map[string]interface{}{
		"uris": opts.uris,
	}
	if opts.position >= 0 {
		jsonBody["position"] = opts.position
	}
	jsonData, err := json.Marshal(jsonBody)
	if err != nil {
		return err
	}
	body := bytes.NewReader(jsonData)

	req, err := newRequest(SpotifyRequestOptions{
		method:      "POST",
		path:        fmt.Sprintf("/playlists/%s/tracks", opts.playlistId),
		accessToken: opts.accessToken,
		body:        body,
	})
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to add items to playlist")
	}

	return nil
}

type GetPlaylistRequestOptions struct {
	accessToken AccessToken
	playlistId  string
}

func getPlaylist(opts GetPlaylistRequestOptions) (*Playlist, error) {
	var playlist Playlist

	if opts.playlistId == "" {
		return &playlist, errors.New("playlist ID is required")
	}

	req, err := newRequest(SpotifyRequestOptions{
		method:      "GET",
		path:        fmt.Sprintf("/playlists/%s", opts.playlistId),
		accessToken: opts.accessToken,
	})

	if err != nil {
		return &playlist, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return &playlist, err
	}

	if resp.StatusCode != http.StatusOK {
		return &playlist, fmt.Errorf("failed to get playlist")
	}

	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&playlist)
	if err != nil {
		return &playlist, err
	}

	return &playlist, nil
}

type UnfollowPlaylistRequestOptions struct {
	accessToken AccessToken
	playlistId  string
}

func unfollowPlaylist(opts UnfollowPlaylistRequestOptions) error {
	if opts.playlistId == "" {
		return errors.New("playlist ID is required")
	}

	req, err := newRequest(SpotifyRequestOptions{
		method:      "DELETE",
		path:        fmt.Sprintf("/playlists/%s/followers", opts.playlistId),
		accessToken: opts.accessToken,
	})
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to unfollow playlist")
	}

	return nil
}
