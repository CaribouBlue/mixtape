package core

import (
	"errors"
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"github.com/CaribouBlue/mixtape/internal/utils"
)

type SessionPhase string

const (
	SubmissionPhase SessionPhase = "submission"
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
	Id                      int64
	Name                    string
	CreatedBy               int64
	CreatedAt               time.Time
	MaxSubmissions          int
	StartAt                 time.Time
	SubmissionPhaseDuration time.Duration
	VotePhaseDuration       time.Duration
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
	Id          int64
	SessionId   int64
	NominatorId int64
	TrackId     string
	Votes       int
}

type VoteEntity struct {
	SessionId   int64
	CandidateId int64
	VoterId     int64
}

type PlayerEntity struct {
	SessionId  int64
	PlayerId   int64
	PlaylistId string
}

type SessionDto struct {
	SessionEntity
	SubmittedCandidates *[]CandidateDto
	BallotCandidates    *[]CandidateDto
	Results             *[]CandidateDto
	CurrentPlayer       *PlayerDto
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
	CandidateEntity
	Track     *TrackEntity
	Vote      *VoteEntity
	Nominator *UserEntity
	Place     int
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

type PlayerDto struct {
	PlayerEntity
	PlaylistUrl string
	DisplayName string
}

func (p *PlayerDto) IsJoinedSession() bool {
	return p.PlayerEntity.PlayerId != 0
}

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

	AddPlayer(sessionId int64, player *PlayerEntity) (*PlayerEntity, error)
	GetPlayer(sessionId int64, playerId int64) (*PlayerEntity, error)
	UpdatePlayerPlaylist(sessionId int64, playerId int64, playlistId string) error
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
	if err != nil {
		return nil, err
	}

	_, err = s.sessionRepository.AddPlayer(session.Id, &PlayerEntity{
		SessionId: session.Id,
		PlayerId:  session.CreatedBy,
	})
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s *SessionService) GetSessionsListForUser(userId int64) (*[]SessionDto, error) {
	sessionEntities, err := s.sessionRepository.GetAllSessions()
	if err != nil {
		return nil, err
	}

	sessions := make([]SessionDto, len(*sessionEntities))

	for i, sessionEntity := range *sessionEntities {
		session := SessionDto{
			SessionEntity: sessionEntity,
			CurrentPlayer: &PlayerDto{},
		}

		player, err := s.sessionRepository.GetPlayer(sessionEntity.Id, userId)
		if err != nil {
			return nil, err
		} else if player != nil {
			session.CurrentPlayer.PlayerEntity = *player
		}

		fmt.Println("Session player ID: ", session.CurrentPlayer.PlayerId)

		sessions[i] = session
	}

	return &sessions, nil
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
	currentPlayer := PlayerDto{}

	currentPlayerEntity, err := s.sessionRepository.GetPlayer(sessionId, userId)
	if err != nil {
		return nil, err
	}

	if currentPlayerEntity != nil {
		currentPlayer.PlayerEntity = *currentPlayerEntity
	}

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
		if currentPlayer.IsJoinedSession() {
			log.Default().Println("Current player playlist ID: ", currentPlayer.PlaylistId)
			if currentPlayer.PlaylistId == "" {
				candidates, err := s.sessionRepository.GetAllCandidates(sessionId)
				if err != nil {
					return nil, err
				}

				trackIds := make([]string, len(*candidates))
				for i, candidate := range *candidates {
					trackIds[i] = candidate.TrackId
				}

				playlistName := fmt.Sprintf("Mixtape: %s %s", session.Name, session.CreatedAt.Format("02-01-06"))
				playlistDetails, err := s.musicService.musicRepository.CreatePlaylist(playlistName, trackIds)
				if err != nil {
					return nil, err
				}

				log.Default().Println("Playlist details: ", playlistDetails.Id, playlistDetails)

				err = s.sessionRepository.UpdatePlayerPlaylist(sessionId, userId, playlistDetails.Id)
				if err != nil {
					return nil, err
				}
				currentPlayer.PlaylistId = playlistDetails.Id
				currentPlayer.PlaylistUrl = playlistDetails.Url
			} else {
				playlistDetails, err := s.musicService.GetPlaylistById(currentPlayer.PlaylistId)
				if err != nil {
					return nil, err
				}
				currentPlayer.PlaylistUrl = playlistDetails.Url
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

			if user, ok := userCache[result.NominatorId]; ok {
				result.Nominator = user
			} else {
				user, err := s.userService.GetUserById(result.NominatorId)
				if err != nil {
					return nil, err
				}
				userCache[result.NominatorId] = user
				result.Nominator = user
			}

			results[i] = result
		}

	}

	sessionView := &SessionDto{
		SessionEntity:       *session,
		SubmittedCandidates: &submittedCandidates,
		BallotCandidates:    &ballotCandidates,
		Results:             &results,
		CurrentPlayer:       &currentPlayer,
	}

	return sessionView, nil
}

func (s *SessionService) JoinSession(sessionId, userId int64) (*PlayerDto, error) {
	player, err := s.sessionRepository.AddPlayer(sessionId, &PlayerEntity{
		SessionId: sessionId,
		PlayerId:  userId,
	})
	if err != nil {
		return nil, err
	}

	return &PlayerDto{
		PlayerEntity: *player,
	}, nil
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
		SessionId:   sessionId,
		NominatorId: userId,
		TrackId:     trackId,
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
		VoterId:     userId,
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
