package spotify

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	netUrl "net/url"
	"strconv"
	"strings"
)

type ItemType string

const (
	AlbumItemType     ItemType = "album"
	ArtistItemType    ItemType = "artist"
	PlaylistItemType  ItemType = "playlist"
	TrackItemType     ItemType = "track"
	ShowItemType      ItemType = "show"
	EpisodeItemType   ItemType = "episode"
	AudioBookItemType ItemType = "audiobook"
)

type includeExternalType string

const (
	IncludeExternalAudio includeExternalType = "audio"
)

type GetSearchResultRequestOptions struct {
	accessToken     AccessToken
	query           string
	itemTypes       []ItemType
	market          string
	limit           int
	offset          int
	includeExternal includeExternalType
}

type SearchResultArtist struct {
	ExternalUrls struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Href string `json:"href"`
	Id   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Uri  string `json:"uri"`
}

type SearchResult struct {
	Tracks struct {
		Href     string `json:"href"`
		Limit    int    `json:"limit"`
		Next     string `json:"next"`
		Offset   int    `json:"offset"`
		Previous string `json:"previous"`
		Total    int    `json:"total"`
		Items    []struct {
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
			Artists          []SearchResultArtist `json:"artists"`
			AvailableMarkets []string             `json:"available_markets"`
			DiscNumber       int                  `json:"disc_number"`
			DurationMs       int                  `json:"duration_ms"`
			Explicit         bool                 `json:"explicit"`
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
		} `json:"items"`
	} `json:"tracks"`
	Artists struct {
		Href     string `json:"href"`
		Limit    int    `json:"limit"`
		Next     string `json:"next"`
		Offset   int    `json:"offset"`
		Previous string `json:"previous"`
		Total    int    `json:"total"`
		Items    []struct {
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Followers struct {
				Href  string `json:"href"`
				Total int    `json:"total"`
			} `json:"followers"`
			Genres []string `json:"genres"`
			Href   string   `json:"href"`
			Id     string   `json:"id"`
			Images []struct {
				Url    string `json:"url"`
				Height int    `json:"height"`
				Width  int    `json:"width"`
			} `json:"images"`
			Name       string `json:"name"`
			Popularity int    `json:"popularity"`
			Type       string `json:"type"`
			Uri        string `json:"uri"`
		} `json:"items"`
	} `json:"artists"`
	Albums struct {
		Href     string `json:"href"`
		Limit    int    `json:"limit"`
		Next     string `json:"next"`
		Offset   int    `json:"offset"`
		Previous string `json:"previous"`
		Total    int    `json:"total"`
		Items    []struct {
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
		} `json:"items"`
	} `json:"albums"`
	Playlists struct {
		Href     string `json:"href"`
		Limit    int    `json:"limit"`
		Next     string `json:"next"`
		Offset   int    `json:"offset"`
		Previous string `json:"previous"`
		Total    int    `json:"total"`
		Items    []struct {
			Collaborative bool   `json:"collaborative"`
			Description   string `json:"description"`
			ExternalUrls  struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
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
				Href  string `json:"href"`
				Total int    `json:"total"`
			} `json:"tracks"`
			Type string `json:"type"`
			Uri  string `json:"uri"`
		} `json:"items"`
	} `json:"playlists"`
	Shows struct {
		Href     string `json:"href"`
		Limit    int    `json:"limit"`
		Next     string `json:"next"`
		Offset   int    `json:"offset"`
		Previous string `json:"previous"`
		Total    int    `json:"total"`
		Items    []struct {
			AvailableMarkets []string `json:"available_markets"`
			Copyrights       []struct {
				Text string `json:"text"`
				Type string `json:"type"`
			} `json:"copyrights"`
			Description     string `json:"description"`
			HtmlDescription string `json:"html_description"`
			Explicit        bool   `json:"explicit"`
			ExternalUrls    struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href   string `json:"href"`
			Id     string `json:"id"`
			Images []struct {
				Url    string `json:"url"`
				Height int    `json:"height"`
				Width  int    `json:"width"`
			} `json:"images"`
			IsExternallyHosted bool     `json:"is_externally_hosted"`
			Languages          []string `json:"languages"`
			MediaType          string   `json:"media_type"`
			Name               string   `json:"name"`
			Publisher          string   `json:"publisher"`
			Type               string   `json:"type"`
			Uri                string   `json:"uri"`
			TotalEpisodes      int      `json:"total_episodes"`
		} `json:"items"`
	} `json:"shows"`
	Episodes struct {
		Href     string `json:"href"`
		Limit    int    `json:"limit"`
		Next     string `json:"next"`
		Offset   int    `json:"offset"`
		Previous string `json:"previous"`
		Total    int    `json:"total"`
		Items    []struct {
			AudioPreviewUrl string `json:"audio_preview_url"`
			Description     string `json:"description"`
			HtmlDescription string `json:"html_description"`
			DurationMs      int    `json:"duration_ms"`
			Explicit        bool   `json:"explicit"`
			ExternalUrls    struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href   string `json:"href"`
			Id     string `json:"id"`
			Images []struct {
				Url    string `json:"url"`
				Height int    `json:"height"`
				Width  int    `json:"width"`
			} `json:"images"`
			IsExternallyHosted   bool     `json:"is_externally_hosted"`
			IsPlayable           bool     `json:"is_playable"`
			Language             string   `json:"language"`
			Languages            []string `json:"languages"`
			Name                 string   `json:"name"`
			ReleaseDate          string   `json:"release_date"`
			ReleaseDatePrecision string   `json:"release_date_precision"`
			ResumePoint          struct {
				FullyPlayed      bool `json:"fully_played"`
				ResumePositionMs int  `json:"resume_position_ms"`
			} `json:"resume_point"`
			Type         string `json:"type"`
			Uri          string `json:"uri"`
			Restrictions struct {
				Reason string `json:"reason"`
			} `json:"restrictions"`
		} `json:"items"`
	} `json:"episodes"`
	Audiobooks struct {
		Href     string `json:"href"`
		Limit    int    `json:"limit"`
		Next     string `json:"next"`
		Offset   int    `json:"offset"`
		Previous string `json:"previous"`
		Total    int    `json:"total"`
		Items    []struct {
			Authors []struct {
				Name string `json:"name"`
			} `json:"authors"`
			AvailableMarkets []string `json:"available_markets"`
			Copyrights       []struct {
				Text string `json:"text"`
				Type string `json:"type"`
			} `json:"copyrights"`
			Description     string `json:"description"`
			HtmlDescription string `json:"html_description"`
			Edition         string `json:"edition"`
			Explicit        bool   `json:"explicit"`
			ExternalUrls    struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href   string `json:"href"`
			Id     string `json:"id"`
			Images []struct {
				Url    string `json:"url"`
				Height int    `json:"height"`
				Width  int    `json:"width"`
			} `json:"images"`
			Languages []string `json:"languages"`
			MediaType string   `json:"media_type"`
			Name      string   `json:"name"`
			Narrators []struct {
				Name string `json:"name"`
			} `json:"narrators"`
			Publisher     string `json:"publisher"`
			Type          string `json:"type"`
			Uri           string `json:"uri"`
			TotalChapters int    `json:"total_chapters"`
		} `json:"items"`
	} `json:"audiobooks"`
}

func getSearchResult(opts GetSearchResultRequestOptions) (*SearchResult, error) {
	var searchRequest SearchResult

	params := netUrl.Values{}
	if opts.query == "" {
		return nil, errors.New("query is required")
	}
	params.Add("q", opts.query)

	if len(opts.itemTypes) == 0 {
		return nil, errors.New("itemTypes is required")
	}
	itemTypes := make([]string, len(opts.itemTypes))
	for i, itemType := range opts.itemTypes {
		itemTypes[i] = string(itemType)
	}
	params.Add("type", strings.Join(itemTypes, ","))

	if opts.market != "" {
		params.Add("market", opts.market)
	}

	if opts.limit != 0 {
		params.Add("limit", strconv.Itoa(opts.limit))
	}

	if opts.offset != 0 {
		params.Add("offset", strconv.Itoa(opts.offset))
	}

	if opts.includeExternal != "" {
		params.Add("include_external", string(opts.includeExternal))
	}

	req, err := newRequest(SpotifyRequestOptions{
		method:      "GET",
		path:        "/search?" + params.Encode(),
		accessToken: opts.accessToken,
	})
	if err != nil {
		return &searchRequest, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return &searchRequest, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		var body string
		if err != nil {
			body = ""
		}
		body = string(bodyBytes)
		return &searchRequest, fmt.Errorf("failed to search Spotify with %s: %s", resp.Status, body)
	}

	err = json.NewDecoder(resp.Body).Decode(&searchRequest)
	if err != nil {
		return &searchRequest, err
	}

	return &searchRequest, nil
}
