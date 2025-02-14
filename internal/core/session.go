package core

import (
	"errors"
	"math"
	"sort"
	"time"
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

type SubmissionEntity struct {
	Id        int64  `json:"id"`
	SessionId int64  `json:"sessionId"`
	UserId    int64  `json:"userId"`
	TrackId   string `json:"trackId"`
}

type VoteEntity struct {
	SessionId    int64 `json:"sessionId"`
	SubmissionId int64 `json:"submissionId"`
	UserId       int64 `json:"userId"`
}

type SessionPlaylistEntity struct {
	SessionId  int64  `json:"sessionId"`
	UserId     int64  `json:"userId"`
	PlaylistId string `json:"playlistId"`
}

type SessionViewDto struct {
	Session         *SessionEntity      `json:"session"`
	UserSubmissions *[]SubmissionEntity `json:"userSubmissions"`
	UserCandidates  *[]CandidateDto     `json:"candidates"`
	Results         *ResultsDto         `json:"results"`
}

type CandidateDto struct {
	Submission *SubmissionEntity `json:"submission"`
	Track      *TrackEntity      `json:"track"`
	Vote       *VoteEntity       `json:"vote"`
}

func NewCandidateDto(sessionId int64) *CandidateDto {
	return &CandidateDto{
		Submission: &SubmissionEntity{
			SessionId: sessionId,
		},
		Vote: &VoteEntity{
			SessionId: sessionId,
		},
	}
}

func (c *CandidateDto) HasVote() bool {
	return c.Vote != nil
}

type ResultDto struct {
	Submission SubmissionEntity `json:"submission"`
	Place      int              `json:"place"`
	Score      int              `json:"score"`
}

type ByScoreCountDesc []ResultDto

func (a ByScoreCountDesc) Len() int           { return len(a) }
func (a ByScoreCountDesc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByScoreCountDesc) Less(i, j int) bool { return a[i].Score > a[j].Score }

type ResultsDto []ResultDto

func NewResultsDto() ResultsDto {
	return []ResultDto{}
}

type SessionRepository interface {
	CreateSession(session *SessionEntity) (*SessionEntity, error)
	GetSessionById(id int64) (*SessionEntity, error)
	GetAllSessions() (*[]SessionEntity, error)
	UpdateSession(session *SessionEntity) (*SessionEntity, error)
	DeleteSession(id int64) error

	AddSubmission(sessionId int64, submission *SubmissionEntity) (*SubmissionEntity, error)
	GetAllSubmissions(sessionId int64) (*[]SubmissionEntity, error)
	GetSubmissionsByUserId(sessionId int64, userId int64) (*[]SubmissionEntity, error)
	GetSubmissionById(sessionId int64, submissionId int64) (*SubmissionEntity, error)
	DeleteSubmission(sessionId int64, submissionId int64) error

	AddVote(sessionId int64, vote *VoteEntity) (*VoteEntity, error)
	GetAllVotes(sessionId int64) (*[]VoteEntity, error)
	GetVotesByUserId(sessionId int64, userId int64) (*[]VoteEntity, error)
	DeleteVote(sessionId int64, userId int64, submissionId int64) error

	GetUserCandidates(sessionId int64, userId int64) (*[]CandidateDto, error)
	GetCandidate(sessionId int64, userId int64, submissionId int64) (*CandidateDto, error)
	GetCandidatesWithVotes(sessionId int64) (*[]CandidateDto, error)

	AddPlaylist(sessionId int64, playlist *SessionPlaylistEntity) (*SessionPlaylistEntity, error)
	FindPlaylist(sessionId int64, userId int64) (*SessionPlaylistEntity, error)
	DeletePlaylist(sessionId int64, userId int64) error
}

type SessionService struct {
	sessionRepository SessionRepository
	musicService      *MusicService
}

func NewSessionService(sessionRepository SessionRepository, musicService *MusicService) *SessionService {
	return &SessionService{
		sessionRepository: sessionRepository,
		musicService:      musicService,
	}
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

func (s *SessionService) GetSession(id int64) (*SessionEntity, error) {
	session, err := s.sessionRepository.GetSessionById(id)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s *SessionService) GetSessionView(sessionId, userId int64) (*SessionViewDto, error) {
	session, err := s.sessionRepository.GetSessionById(sessionId)
	if err != nil {
		return nil, err
	}

	submissions, err := s.sessionRepository.GetSubmissionsByUserId(sessionId, userId)
	if err != nil {
		return nil, err
	}

	candidates, err := s.sessionRepository.GetUserCandidates(sessionId, userId)
	if err != nil {
		return nil, err
	}

	results, err := s.GetResults(sessionId)
	if err != nil {
		return nil, err
	}

	sessionView := &SessionViewDto{
		Session:         session,
		UserSubmissions: submissions,
		UserCandidates:  candidates,
		Results:         results,
	}

	return sessionView, nil
}

func (s *SessionService) SearchSubmissionTracks(sessionId int64, query string) (*[]CandidateDto, error) {
	tracks, err := s.musicService.SearchTracks(query)
	if err != nil {
		return nil, err
	}

	candidates := make([]CandidateDto, 0)
	for _, track := range tracks {
		candidate := NewCandidateDto(sessionId)
		candidate.Submission.TrackId = track.Id
		candidate.Track = &track
		candidates = append(candidates, *candidate)
	}

	return &candidates, nil
}

func (s *SessionService) AddUserSubmission(sessionId, userId int64, trackId string) (*SubmissionEntity, error) {
	// TODO: improve validation logic
	session, err := s.sessionRepository.GetSessionById(sessionId)
	if err != nil {
		return nil, err
	}

	submissions, err := s.sessionRepository.GetSubmissionsByUserId(sessionId, userId)
	if err != nil {
		return nil, err
	}

	if session.MaxSubmissions <= len(*submissions) {
		return nil, ErrNoSubmissionsLeft
	}

	for _, submission := range *submissions {
		if submission.TrackId == trackId {
			return nil, ErrDuplicateSubmission
		}
	}

	submission, err := s.sessionRepository.AddSubmission(sessionId, &SubmissionEntity{
		SessionId: sessionId,
		UserId:    userId,
		TrackId:   trackId,
	})
	if err != nil {
		return nil, err
	}

	return submission, nil
}

func (s *SessionService) GetAllSubmissions(sessionId int64) (*[]SubmissionEntity, error) {
	submissions, err := s.sessionRepository.GetAllSubmissions(sessionId)
	if err != nil {
		return nil, err
	}
	return submissions, nil
}

func (s *SessionService) GetUserSubmissions(sessionId, userId int64) (*[]SubmissionEntity, error) {
	submissions, err := s.sessionRepository.GetSubmissionsByUserId(sessionId, userId)
	if err != nil {
		return nil, err
	}

	return submissions, nil
}

func (s *SessionService) GetSubmissionById(sessionId, submissionId int64) (*SubmissionEntity, error) {
	submission, err := s.sessionRepository.GetSubmissionById(sessionId, submissionId)
	if err != nil {
		return nil, err
	}

	return submission, nil
}

func (s *SessionService) RemoveUserSubmission(sessionId, userId, submissionId int64) error {
	err := s.sessionRepository.DeleteSubmission(sessionId, submissionId)
	if err != nil {
		return err
	}

	return nil
}

func (s *SessionService) GetUserCandidates(sessionId, userId int64) (*[]CandidateDto, error) {
	candidates, err := s.sessionRepository.GetUserCandidates(sessionId, userId)
	return candidates, err
}

func (s *SessionService) GetUserCandidate(sessionId, userId, submissionId int64) (*CandidateDto, error) {
	candidate, err := s.sessionRepository.GetCandidate(sessionId, userId, submissionId)
	if err != nil {
		return nil, err
	}

	return candidate, nil
}

func (s *SessionService) VoteForCandidate(sessionId, userId, submissionId int64) (*CandidateDto, error) {
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
		SessionId:    sessionId,
		UserId:       userId,
		SubmissionId: submissionId,
	})
	if err != nil {
		return nil, err
	}

	candidate, err := s.sessionRepository.GetCandidate(sessionId, userId, submissionId)
	if err != nil {
		return nil, err
	}

	return candidate, nil
}

func (s *SessionService) RemoveVoteForCandidate(sessionId, userId, submissionId int64) (*CandidateDto, error) {
	err := s.sessionRepository.DeleteVote(sessionId, userId, submissionId)
	if err != nil {
		return nil, err
	}

	candidate, err := s.sessionRepository.GetCandidate(sessionId, userId, submissionId)
	if err != nil {
		return nil, err
	}

	return candidate, nil
}

func (s *SessionService) AddPlaylist(sessionId, userId int64, playlistId string) (*SessionPlaylistEntity, error) {
	playlist, err := s.sessionRepository.AddPlaylist(sessionId, &SessionPlaylistEntity{
		SessionId:  sessionId,
		UserId:     userId,
		PlaylistId: playlistId,
	})
	if err != nil {
		return nil, err
	}

	return playlist, nil
}

func (s *SessionService) GetUserPlaylist(sessionId, userId int64) (*SessionPlaylistEntity, error) {
	playlist, err := s.sessionRepository.FindPlaylist(sessionId, userId)
	if err != nil {
		return nil, err
	}

	if playlist == nil {
		return nil, ErrPlaylistNotFound
	}

	return playlist, nil
}

func (s *SessionService) GetResults(sessionId int64) (*ResultsDto, error) {
	resultsDto := NewResultsDto()

	candidates, err := s.sessionRepository.GetCandidatesWithVotes(int64(sessionId))
	if err != nil {
		return nil, err
	}

	resultMap := make(map[int64]ResultDto)
	for _, candidate := range *candidates {
		submission := candidate.Submission

		if result, ok := resultMap[submission.Id]; ok {
			result.Score++
		} else {
			resultMap[submission.Id] = ResultDto{
				Submission: *submission,
				Score:      1,
			}
		}

	}

	for _, result := range resultMap {
		resultsDto = append(resultsDto, result)
	}

	sort.Sort(ByScoreCountDesc(resultsDto))

	place := 1
	currentBest := 0
	for i, result := range resultsDto {
		if result.Score < currentBest {
			place += 1
		}

		if result.Score == 0 {
			place = -1
		}

		currentBest = result.Score
		result.Place = place
		resultsDto[i] = result
	}

	return &resultsDto, nil
}
