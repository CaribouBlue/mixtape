package music

import (
	"github.com/CaribouBlue/top-spot/internal/music/spotify"
	"github.com/CaribouBlue/top-spot/internal/user"
)

type MusicService interface {
	Authenticate(user *user.User) error
	SearchTracks(query string) ([]*Track, error)
	GetTrack(trackId string) (*Track, error)
	CreatePlaylist(playlist *Playlist, trackIds []string) error
	GetPlaylist(playlistId string) (*Playlist, error)
}

type spotifyMusicService struct {
	client *spotify.Client
}

func NewSpotifyMusicService() MusicService {
	client := spotify.DefaultClient()
	return &spotifyMusicService{
		client: client,
	}
}

func (s *spotifyMusicService) Authenticate(u *user.User) error {
	s.client.SetAccessToken(u.SpotifyAccessToken)
	_, err := s.client.GetValidAccessToken()
	return err
}

func (s *spotifyMusicService) SearchTracks(query string) ([]*Track, error) {
	searchResult, err := s.client.SearchTracks(query)
	if err != nil {
		return nil, err
	}

	tracks := make([]*Track, len(searchResult.Tracks.Items))
	for i, track := range searchResult.Tracks.Items {
		tracks[i] = &Track{
			Id:      track.Id,
			Name:    track.Name,
			Artists: make([]Artist, len(track.Artists)),
		}

		for j, artist := range track.Artists {
			tracks[i].Artists[j] = Artist{
				Id:   artist.Id,
				Name: artist.Name,
			}
		}
	}

	return tracks, nil
}

func (s *spotifyMusicService) GetTrack(trackId string) (*Track, error) {
	t, err := s.client.GetTrack(trackId)
	if err != nil {
		return nil, err
	}

	track := &Track{
		Id:      t.Id,
		Name:    t.Name,
		Artists: make([]Artist, len(t.Artists)),
	}

	for i, artist := range t.Artists {
		track.Artists[i] = Artist{
			Id:   artist.Id,
			Name: artist.Name,
		}
	}

	return track, nil
}

func (s *spotifyMusicService) CreatePlaylist(playlist *Playlist, trackIds []string) error {
	newPlaylist, err := s.client.CreatePlaylist(playlist.Name)
	if err != nil {
		return err
	}

	playlist.Id = newPlaylist.Id

	return s.client.AddTracksToPlaylist(playlist.Id, trackIds)
}

func (s *spotifyMusicService) GetPlaylist(playlistId string) (*Playlist, error) {
	playlist, err := s.client.GetPlaylist(playlistId)
	if err != nil {
		return nil, err
	}

	return &Playlist{
		Id:   playlist.Id,
		Name: playlist.Name,
		Url:  playlist.ExternalUrls.Spotify,
	}, nil
}
