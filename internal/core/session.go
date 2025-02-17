package core

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/CaribouBlue/top-spot/internal/utils"
)

type SessionPhase string

const (
	SubmissionPhase SessionPhase = "submissions"
	VotePhase       SessionPhase = "voting"
	ResultPhase     SessionPhase = "results"
)

var (
	ErrNoSubmissionsLeft   = errors.New("no submissions left")
	ErrDuplicateSubmission = errors.New("duplicate submission")
	ErrNoVotesLeft         = errors.New("no votes left")
	ErrPlaylistNotFound    = errors.New("playlist not found")
)

type SessionEntity struct {
	Id                      int64         `json:"id"`
	Name                    string        `json:"name"`
	CreatedBy               int64         `json:"createdBy"`
	CreatedAt               time.Time     `json:"createdAt"`
	MaxSubmissions          int           `json:"maxSubmissions"`
	StartAt                 time.Time     `json:"startAt"`
	SubmissionPhaseDuration time.Duration `json:"submissionDuration"`
	VotePhaseDuration       time.Duration `json:"voteDuration"`
}

type SessionOption func(*SessionEntity)

func WithSessionStartAt(startAt time.Time) SessionOption {
	return func(session *SessionEntity) {
		session.StartAt = startAt
	}
}

func WithSubmissionDuration(submissionDuration time.Duration) SessionOption {
	return func(session *SessionEntity) {
		session.SubmissionPhaseDuration = submissionDuration
	}
}

func WithVoteDuration(voteDuration time.Duration) SessionOption {
	return func(session *SessionEntity) {
		session.VotePhaseDuration = voteDuration
	}
}

func NewSessionEntity(name string, createdBy int64, options ...SessionOption) *SessionEntity {
	day := 24 * time.Hour
	now := time.Now()

	session := &SessionEntity{
		Name:                    name,
		CreatedBy:               createdBy,
		CreatedAt:               now,
		StartAt:                 now,
		SubmissionPhaseDuration: day * 5,
		VotePhaseDuration:       day * 5,
		MaxSubmissions:          5,
	}

	for _, opt := range options {
		opt(session)
	}

	return session
}

func (s *SessionEntity) Phase() SessionPhase {
	if time.Since(s.StartAt) < s.SubmissionPhaseDuration {
		return SubmissionPhase
	}

	if time.Since(s.StartAt) < s.SubmissionPhaseDuration+s.VotePhaseDuration {
		return VotePhase
	}

	return ResultPhase
}

func (s *SessionEntity) RemainingPhaseDuration() time.Duration {
	sessionPhase := s.Phase()

	switch {
	case sessionPhase == SubmissionPhase:
		return s.SubmissionPhaseDuration - time.Since(s.StartAt)
	case sessionPhase == VotePhase:
		return s.SubmissionPhaseDuration + s.VotePhaseDuration - time.Since(s.StartAt)
	default:
		return 0
	}
}

func (s *SessionEntity) MaxVotes() int {
	maxVotes := math.Floor(float64(s.MaxSubmissions) / 2)
	if maxVotes > 10 {
		return 10
	} else {
		return int(maxVotes)
	}
}

type CandidateEntity struct {
	Id        int64  `json:"id"`
	SessionId int64  `json:"sessionId"`
	UserId    int64  `json:"userId"`
	TrackId   string `json:"trackId"`
	Votes     int    `json:"votes"`
}

type VoteEntity struct {
	SessionId   int64 `json:"sessionId"`
	CandidateId int64 `json:"candidateId"`
	UserId      int64 `json:"userId"`
}

type SessionPlaylistEntity struct {
	SessionId  int64  `json:"sessionId"`
	UserId     int64  `json:"userId"`
	PlaylistId string `json:"playlistId"`
}

type PlaylistDto struct {
	SessionPlaylistEntity
	Name string
	Url  string
}

type SessionDto struct {
	SessionEntity       `json:"session"`
	SubmittedCandidates *[]CandidateDto `json:"submittedCandidates"`
	BallotCandidates    *[]CandidateDto `json:"ballotCandidates"`
	Results             *[]CandidateDto `json:"results"`
	Playlist            *PlaylistDto    `json:"playlist"`
}

func (s *SessionDto) VoteCount() int {
	return utils.Reduce(*s.BallotCandidates, func(count int, candidate CandidateDto) int {
		if candidate.Vote != nil {
			return count + 1
		}
		return count
	}, 0)
}

type CandidateDto struct {
	CandidateEntity `json:"candidate"`
	Track           *TrackEntity `json:"track"`
	Vote            *VoteEntity  `json:"vote"`
	Owner           *UserEntity  `json:"owner"`
	Place           int          `json:"place"`
}

func NewCandidateDto(sessionId int64) *CandidateDto {
	return &CandidateDto{
		CandidateEntity: CandidateEntity{
			SessionId: sessionId,
		},
		Vote: &VoteEntity{
			SessionId: sessionId,
		},
	}
}

type ByVoteCountDesc []CandidateDto

func (a ByVoteCountDesc) Len() int           { return len(a) }
func (a ByVoteCountDesc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByVoteCountDesc) Less(i, j int) bool { return a[i].Votes > a[j].Votes }

type SessionRepository interface {
	CreateSession(session *SessionEntity) (*SessionEntity, error)
	GetSessionById(id int64) (*SessionEntity, error)
	GetAllSessions() (*[]SessionEntity, error)

	AddCandidate(sessionId int64, candidate *CandidateEntity) (*CandidateEntity, error)
	GetAllCandidates(sessionId int64) (*[]CandidateEntity, error)
	GetCandidatesByUserId(sessionId int64, userId int64) (*[]CandidateEntity, error)
	GetCandidateByNotUserId(sessionId int64, userId int64) (*[]CandidateEntity, error)
	GetCandidateById(sessionId int64, candidateId int64) (*CandidateEntity, error)
	DeleteCandidate(sessionId int64, candidateId int64) error

	AddVote(sessionId int64, vote *VoteEntity) (*VoteEntity, error)
	GetVotesByUserId(sessionId int64, userId int64) (*[]VoteEntity, error)
	GetVote(sessionId int64, userId int64, candidateId int64) (*VoteEntity, error)
	DeleteVote(sessionId int64, userId int64, candidateId int64) error

	AddPlaylist(sessionId int64, playlist *SessionPlaylistEntity) (*SessionPlaylistEntity, error)
	FindPlaylist(sessionId int64, userId int64) (*SessionPlaylistEntity, error)
}

type SessionService struct {
	sessionRepository SessionRepository
	userService       *UserService
	musicService      *MusicService
}

func NewSessionService(sessionRepository SessionRepository, userService *UserService, musicService *MusicService) *SessionService {
	return &SessionService{
		sessionRepository: sessionRepository,
		userService:       userService,
		musicService:      musicService,
	}
}

func (s *SessionService) getCandidateDtoFromEntity(entity *CandidateEntity, userId int64) (*CandidateDto, error) {
	track, err := s.musicService.GetTrackById(entity.TrackId)
	if err != nil {
		return nil, err
	}

	vote, err := s.sessionRepository.GetVote(entity.SessionId, userId, entity.Id)
	if err != nil {
		return nil, err
	}

	dto := &CandidateDto{
		CandidateEntity: *entity,
		Track:           track,
		Vote:            vote,
	}

	return dto, nil
}

func (s *SessionService) CreateSession(session *SessionEntity) (*SessionEntity, error) {
	session, err := s.sessionRepository.CreateSession(session)

	return session, err
}

func (s *SessionService) GetSessionsList() (*[]SessionEntity, error) {
	sessions, err := s.sessionRepository.GetAllSessions()
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

func (s *SessionService) GetSessionData(id int64) (*SessionEntity, error) {
	session, err := s.sessionRepository.GetSessionById(id)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s *SessionService) GetSessionView(sessionId, userId int64) (*SessionDto, error) {
	session, err := s.sessionRepository.GetSessionById(sessionId)
	if err != nil {
		return nil, err
	}

	submittedCandidates := []CandidateDto{}
	ballotCandidates := []CandidateDto{}
	results := []CandidateDto{}
	playlist := &PlaylistDto{}

	if session.Phase() != ResultPhase {
		candidatesSubmittedByUser, err := s.sessionRepository.GetCandidatesByUserId(sessionId, userId)
		if err != nil {
			return nil, err
		}

		for _, candidate := range *candidatesSubmittedByUser {
			candidateDto, err := s.getCandidateDtoFromEntity(&candidate, userId)
			if err != nil {
				return nil, err
			}

			submittedCandidates = append(submittedCandidates, *candidateDto)
		}
	}

	if session.Phase() != SubmissionPhase {
		playlistEntity, err := s.sessionRepository.FindPlaylist(sessionId, userId)
		if err != nil {
			return nil, err
		}

		if playlistEntity == nil {
			candidates, err := s.sessionRepository.GetAllCandidates(sessionId)
			if err != nil {
				return nil, err
			}

			trackIds := make([]string, len(*candidates))
			for i, candidate := range *candidates {
				trackIds[i] = candidate.TrackId
			}

			playlistName := fmt.Sprintf("Top Spot: %s %s", session.Name, session.CreatedAt.Format("02-01-06"))
			playlistDetails, err := s.musicService.musicRepository.CreatePlaylist(playlistName, trackIds)
			if err != nil {
				return nil, err
			}

			playlistEntity, err = s.sessionRepository.AddPlaylist(sessionId, &SessionPlaylistEntity{
				SessionId:  sessionId,
				UserId:     userId,
				PlaylistId: playlistDetails.Id,
			})
			if err != nil {
				return nil, err
			}

			playlist = &PlaylistDto{
				SessionPlaylistEntity: *playlistEntity,
				Name:                  playlistDetails.Name,
				Url:                   playlistDetails.Url,
			}
		} else {
			playlistDetails, err := s.musicService.GetPlaylistById(playlistEntity.PlaylistId)
			if err != nil {
				return nil, err
			}

			playlist = &PlaylistDto{
				SessionPlaylistEntity: *playlistEntity,
				Name:                  playlistDetails.Name,
				Url:                   playlistDetails.Url,
			}
		}
	}

	if session.Phase() == VotePhase {
		candidatesNotSubmittedByUser, err := s.sessionRepository.GetCandidateByNotUserId(sessionId, userId)
		if err != nil {
			return nil, err
		}

		for _, candidate := range *candidatesNotSubmittedByUser {
			candidateDto, err := s.getCandidateDtoFromEntity(&candidate, userId)
			if err != nil {
				return nil, err
			}

			// log.Default().Println("CandidateDto: ", candidateDto)
			// log.Default().Println("Vote: ", candidateDto.Vote)

			ballotCandidates = append(submittedCandidates, *candidateDto)
		}
	}

	if session.Phase() == ResultPhase {
		candidates, err := s.sessionRepository.GetAllCandidates(sessionId)
		if err != nil {
			return nil, err
		}

		for _, candidate := range *candidates {
			candidateDto, err := s.getCandidateDtoFromEntity(&candidate, userId)
			if err != nil {
				return nil, err
			}

			results = append(results, *candidateDto)
		}

		sort.Sort(ByVoteCountDesc(results))

		userCache := make(map[int64]*UserEntity)

		place := 1
		currentBest := results[0].Votes
		for i, result := range results {
			if result.Votes < currentBest {
				place += 1
			}

			if result.Votes == 0 {
				place = -1
			}

			currentBest = result.Votes
			result.Place = place

			if user, ok := userCache[result.UserId]; ok {
				result.Owner = user
			} else {
				user, err := s.userService.GetUserById(result.UserId)
				if err != nil {
					return nil, err
				}
				userCache[result.UserId] = user
				result.Owner = user
			}

			results[i] = result
		}

	}

	sessionView := &SessionDto{
		SessionEntity:       *session,
		SubmittedCandidates: &submittedCandidates,
		BallotCandidates:    &ballotCandidates,
		Results:             &results,
		Playlist:            playlist,
	}

	return sessionView, nil
}

func (s *SessionService) SearchCandidateSubmissions(sessionId int64, query string) (*[]CandidateDto, error) {
	tracks, err := s.musicService.SearchTracks(query)
	if err != nil {
		return nil, err
	}

	candidates := make([]CandidateDto, 0)
	for _, track := range tracks {
		candidate := NewCandidateDto(sessionId)
		candidate.TrackId = track.Id
		candidate.Track = &track
		candidates = append(candidates, *candidate)
	}

	return &candidates, nil
}

func (s *SessionService) SubmitCandidate(sessionId, userId int64, trackId string) (*CandidateDto, error) {
	// TODO: improve validation logic
	session, err := s.sessionRepository.GetSessionById(sessionId)
	if err != nil {
		return nil, err
	}

	candidates, err := s.sessionRepository.GetCandidatesByUserId(sessionId, userId)
	if err != nil {
		return nil, err
	}

	if session.MaxSubmissions <= len(*candidates) {
		return nil, ErrNoSubmissionsLeft
	}

	for _, candidate := range *candidates {
		if candidate.TrackId == trackId {
			return nil, ErrDuplicateSubmission
		}
	}

	candidate, err := s.sessionRepository.AddCandidate(sessionId, &CandidateEntity{
		SessionId: sessionId,
		UserId:    userId,
		TrackId:   trackId,
	})
	if err != nil {
		return nil, err
	}

	candidateDto, err := s.getCandidateDtoFromEntity(candidate, userId)
	if err != nil {
		return nil, err
	}

	return candidateDto, nil
}

func (s *SessionService) RemoveCandidate(sessionId, userId, candidateId int64) error {
	err := s.sessionRepository.DeleteCandidate(sessionId, candidateId)
	if err != nil {
		return err
	}

	return nil
}

func (s *SessionService) VoteForCandidate(sessionId, userId, candidateId int64) (*CandidateDto, error) {
	// TODO: improve validation logic

	session, err := s.sessionRepository.GetSessionById(sessionId)
	if err != nil {
		return nil, err
	}

	votes, err := s.sessionRepository.GetVotesByUserId(sessionId, userId)
	if err != nil {
		return nil, err
	}

	if len(*votes) >= session.MaxVotes() {
		return nil, ErrNoVotesLeft
	}

	_, err = s.sessionRepository.AddVote(sessionId, &VoteEntity{
		SessionId:   sessionId,
		UserId:      userId,
		CandidateId: candidateId,
	})
	if err != nil {
		return nil, err
	}

	candidate, err := s.sessionRepository.GetCandidateById(sessionId, candidateId)
	if err != nil {
		return nil, err
	}

	candidateDto, err := s.getCandidateDtoFromEntity(candidate, userId)
	if err != nil {
		return nil, err
	}

	return candidateDto, nil
}

func (s *SessionService) RemoveVoteForCandidate(sessionId, userId, candidateId int64) (*CandidateDto, error) {
	err := s.sessionRepository.DeleteVote(sessionId, userId, candidateId)
	if err != nil {
		return nil, err
	}

	candidate, err := s.sessionRepository.GetCandidateById(sessionId, candidateId)
	if err != nil {
		return nil, err
	}

	candidateDto, err := s.getCandidateDtoFromEntity(candidate, userId)
	if err != nil {
		return nil, err
	}

	return candidateDto, nil
}
