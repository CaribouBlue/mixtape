package spotify

import (
	"fmt"
	"io"
	"net/http"
)

type SpotifyRequestOptions struct {
	accessToken AccessToken
	method      string
	path        string
	body        io.Reader
}

func newRequest(opts SpotifyRequestOptions) (*http.Request, error) {
	url := fmt.Sprintf("https://api.spotify.com/v1%s", opts.path)
	req, err := http.NewRequest(opts.method, url, opts.body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+opts.accessToken.AccessToken)

	return req, nil
}
