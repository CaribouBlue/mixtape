package session

import (
	"errors"

	"github.com/CaribouBlue/top-spot/internal/entities/user"
	"github.com/google/uuid"
)

var (
	ErrSubmissionNotFound = errors.New("submission not found")
	ErrVoteNotFound       = errors.New("vote not found")
	ErrVoteExists         = errors.New("vote already exists")
	ErrPlaylistNotFound   = errors.New("playlist not found")
	ErrPlaylistExists     = errors.New("playlist already exists")
	ErrResultNotFound     = errors.New("result not found")
)

type SessionService interface {
	GetAll() ([]*Session, error)
	GetOne(sessionId int64) (*Session, error)
	Create(creator *user.User, name string) (*Session, error)
	Update(session *Session) error

	GetSubmissions(sessionId int64) ([]Submission, error)
	GetSubmission(sessionId int64, submissionId string) (*Submission, error)
	AddSubmission(sessionId int64, submission *Submission) (*Session, error)
	RemoveSubmission(sessionId int64, submissionId string) (*Session, error)

	GetVotes(sessionId int64) ([]Vote, error)
	GetVote(sessionId int64, voteId string) (*Vote, error)
	AddVote(sessionId int64, vote *Vote) (*Session, error)
	RemoveVote(sessionId int64, voteId string) (*Session, error)

	GetPlaylist(sessionId int64, userId int64) (*Playlist, error)
	AddPlaylist(sessionId int64, playlistId string, userId int64) (*Session, error)

	GetResults(sessionId int64) (*[]Result, error)
	GetResult(sessionId int64, resultId string) (*Result, error)
}

type sessionService struct {
	repo           SessionRepo
	collectionName string
}

func NewSessionService(repo SessionRepo) SessionService {
	return &sessionService{
		repo:           repo,
		collectionName: "session",
	}
}

func (s *sessionService) GetAll() ([]*Session, error) {
	return s.repo.GetSessions()
}

func (s *sessionService) GetOne(sessionId int64) (*Session, error) {
	session, err := s.repo.GetSession(sessionId)
	if err != nil {
		return nil, err
	}

	if session.Phase() == ResultPhase && len(session.Results) == 0 {
		results := session.DeriveResults()
		session.Results = results
		err = s.repo.UpdateSession(session)
		if err != nil {
			return nil, err
		}
	}

	return session, nil
}

func (s *sessionService) Create(creator *user.User, name string) (*Session, error) {
	if name == "" {
		return nil, errors.New("session name cannot be empty")
	}

	newSession := NewSession(name)
	newSession.CreatedBy = creator.Id
	return newSession, s.repo.CreateSession(newSession)
}

func (s *sessionService) Update(session *Session) error {
	return s.repo.UpdateSession(session)
}

func (s *sessionService) GetSubmissions(sessionId int64) ([]Submission, error) {
	session, err := s.repo.GetSession(sessionId)
	if err != nil {
		return nil, err
	}

	return session.Submissions, nil
}

func (s *sessionService) GetSubmission(sessionId int64, submissionId string) (*Submission, error) {
	session, err := s.repo.GetSession(sessionId)
	if err != nil {
		return nil, err
	}

	for _, submission := range session.Submissions {
		if submission.Id == submissionId {
			return &submission, nil
		}
	}

	return nil, ErrSubmissionNotFound
}

func (s *sessionService) AddSubmission(sessionId int64, submission *Submission) (*Session, error) {
	session, err := s.repo.GetSession(sessionId)
	if err != nil {
		return nil, err
	}

	submission.Id = uuid.New().String()

	session.Submissions = append(session.Submissions, *submission)

	err = s.repo.UpdateSession(session)

	return session, err
}

func (s *sessionService) RemoveSubmission(sessionId int64, submissionId string) (*Session, error) {
	session, err := s.repo.GetSession(sessionId)
	if err != nil {
		return nil, err
	}

	for i, submission := range session.Submissions {
		if submission.Id == submissionId {
			session.Submissions = append(session.Submissions[:i], session.Submissions[i+1:]...)
			return session, s.repo.UpdateSession(session)
		}
	}

	return nil, ErrSubmissionNotFound
}

func (s *sessionService) GetVotes(sessionId int64) ([]Vote, error) {
	session, err := s.repo.GetSession(sessionId)
	if err != nil {
		return nil, err
	}

	return session.Votes, nil
}

func (s *sessionService) GetVote(sessionId int64, voteId string) (*Vote, error) {
	session, err := s.repo.GetSession(sessionId)
	if err != nil {
		return nil, err
	}

	for _, vote := range session.Votes {
		if vote.Id == voteId {
			return &vote, nil
		}
	}

	return nil, ErrVoteNotFound
}

func (s *sessionService) AddVote(sessionId int64, vote *Vote) (*Session, error) {
	session, err := s.repo.GetSession(sessionId)
	if err != nil {
		return nil, err
	}

	for _, v := range session.Votes {
		if v.UserId == vote.UserId && v.SubmissionId == vote.SubmissionId {
			return nil, ErrVoteExists
		}
	}

	vote.Id = uuid.New().String()

	session.Votes = append(session.Votes, *vote)

	return session, s.repo.UpdateSession(session)
}

func (s *sessionService) RemoveVote(sessionId int64, voteId string) (*Session, error) {
	session, err := s.repo.GetSession(sessionId)
	if err != nil {
		return nil, err
	}

	for i, vote := range session.Votes {
		if vote.Id == voteId {
			session.Votes = append(session.Votes[:i], session.Votes[i+1:]...)
			return session, s.repo.UpdateSession(session)
		}
	}

	return nil, ErrVoteNotFound
}

func (s *sessionService) GetPlaylist(sessionId int64, userId int64) (*Playlist, error) {
	session, err := s.repo.GetSession(sessionId)
	if err != nil {
		return nil, err
	}

	for _, playlist := range session.Playlists {
		if playlist.UserId == userId {
			return &playlist, nil
		}
	}

	return nil, ErrPlaylistNotFound
}

func (s *sessionService) AddPlaylist(sessionId int64, playlistId string, userId int64) (*Session, error) {
	session, err := s.repo.GetSession(sessionId)
	if err != nil {
		return nil, err
	}

	for _, playlist := range session.Playlists {
		if playlist.UserId == userId {
			return nil, ErrPlaylistExists
		}
	}

	session.Playlists = append(session.Playlists, Playlist{
		Id:     playlistId,
		UserId: userId,
	})

	return session, s.repo.UpdateSession(session)
}

func (s *sessionService) GetResults(sessionId int64) (*[]Result, error) {
	session, err := s.repo.GetSession(sessionId)
	if err != nil {
		return nil, err
	}

	if session.Phase() == ResultPhase && len(session.Results) == 0 {
		results := session.DeriveResults()
		session.Results = results
		err = s.repo.UpdateSession(session)
		if err != nil {
			return nil, err
		}
	}

	return &session.Results, nil
}

func (s *sessionService) GetResult(sessionId int64, resultId string) (*Result, error) {
	session, err := s.repo.GetSession(sessionId)
	if err != nil {
		return nil, err
	}

	for _, result := range session.Results {
		if result.Id == resultId {
			return &result, nil
		}
	}

	return nil, ErrResultNotFound
}
