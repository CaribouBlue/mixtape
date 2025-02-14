package core

type TrackEntity struct {
	Id       string
	Name     string
	Artists  []ArtistEntity
	Album    AlbumEntity
	Explicit bool
	Url      string
}

type ArtistEntity struct {
	Id   string
	Name string
	Url  string
}

type AlbumEntity struct {
	Id   string
	Name string
	Url  string
}

type PlaylistEntity struct {
	Id   string
	Name string
	Url  string
}

type MusicRepository interface {
	AuthenticateUser(user *UserEntity) error

	GetTrackById(trackId string) (*TrackEntity, error)
	SearchTracks(query string) ([]TrackEntity, error)

	CreatePlaylist(name string, trackIds []string) (*PlaylistEntity, error)
	GetPlaylistById(playlistId string) (*PlaylistEntity, error)
}

type MusicService struct {
	musicRepository MusicRepository
}

func NewMusicService(trackRepository MusicRepository) *MusicService {
	return &MusicService{
		musicRepository: trackRepository,
	}
}

func (s *MusicService) Authenticate(user *UserEntity) error {
	err := s.musicRepository.AuthenticateUser(user)
	if err != nil {
		return err
	}
	return nil
}

func (s *MusicService) GetTrackById(trackId string) (*TrackEntity, error) {
	return s.musicRepository.GetTrackById(trackId)
}

func (s *MusicService) SearchTracks(query string) ([]TrackEntity, error) {
	tracks, err := s.musicRepository.SearchTracks(query)
	if err != nil {
		return nil, err
	}
	return tracks, nil
}

func (s *MusicService) CreatePlaylist(name string, trackIds []string) (*PlaylistEntity, error) {
	playlist, err := s.musicRepository.CreatePlaylist(name, trackIds)
	if err != nil {
		return nil, err
	}
	return playlist, nil
}

func (s *MusicService) GetPlaylistById(playlistId string) (*PlaylistEntity, error) {
	playlist, err := s.musicRepository.GetPlaylistById(playlistId)
	if err != nil {
		return nil, err
	}
	return playlist, nil
}
