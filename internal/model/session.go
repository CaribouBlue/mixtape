package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/CaribouBlue/top-spot/internal/db"
	"github.com/google/uuid"
)

type SessionPhase string

const (
	SubmissionPhase SessionPhase = "submission"
	VotePhase       SessionPhase = "vote"
	ResultPhase     SessionPhase = "result"
)

var (
	ErrSubmissionsMaxedOut          = errors.New("user has already submitted the maximum number of submissions")
	ErrSubmissionNotFound           = errors.New("no submission found with the given ID")
	ErrSubmissionUpdateUnauthorized = errors.New("user is unauthorized to update submission")

	ErrVoteNotFound           = errors.New("no vote found with the given ID")
	ErrVoteUpdateUnauthorized = errors.New("user is unauthorized to update vote")
)

type SubmissionData struct {
	Id      string `json:"id"`
	UserId  int64  `json:"userId"`
	TrackId string `json:"trackId"`
}

func NewSubmissionData(userId int64, trackId string) *SubmissionData {
	return &SubmissionData{
		Id:      uuid.New().String(),
		UserId:  userId,
		TrackId: trackId,
	}
}

type VoteData struct {
	Id           string `json:"id"`
	UserId       int64  `json:"userId"`
	SubmissionId string `json:"submissionId"`
}

func NewVoteModel(userId int64, submissionId string) *VoteData {
	return &VoteData{
		Id:           uuid.New().String(),
		UserId:       userId,
		SubmissionId: submissionId,
	}
}

type SessionData struct {
	Id                 int64            `json:"id"`
	Name               string           `json:"name"`
	Submissions        []SubmissionData `json:"submissions"`
	Votes              []VoteData       `json:"votes"`
	CreatedAt          time.Time        `json:"createdAt"`
	MaxSubmissions     int              `json:"maxSubmissions"`
	StartAt            time.Time        `json:"startAt"`
	SubmissionDuration time.Duration    `json:"submissionDuration"`
	VoteDuration       time.Duration    `json:"voteDuration"`
}

type SessionModel struct {
	Data SessionData
	db   db.Db
}

func NewSessionModel(db db.Db, opts ...OptsFn) *SessionModel {
	dayDuration := time.Hour * 24
	data := SessionData{
		Submissions:        make([]SubmissionData, 0),
		Votes:              make([]VoteData, 0),
		CreatedAt:          time.Now(),
		StartAt:            time.Now(),
		MaxSubmissions:     5,
		SubmissionDuration: 5 * dayDuration,
		VoteDuration:       14 * dayDuration,
	}

	model := &SessionModel{
		Data: data,
		db:   db,
	}

	for _, opt := range opts {
		_ = opt(model)
	}

	return model
}

func (session *SessionModel) Name() string {
	return "game_session"
}

func (session *SessionModel) Id() int64 {
	return session.Data.Id
}

func (session *SessionModel) SetId(id int64) {
	session.Data.Id = id
}

func (session *SessionModel) Scan(value interface{}) error {
	return json.Unmarshal([]byte(value.(string)), &session.Data)
}

func (session *SessionModel) Value() (driver.Value, error) {
	return json.Marshal(session.Data)
}

func (session *SessionModel) Create() error {
	return session.db.CreateRecord(session)
}

func (session SessionModel) ReadAll() ([]*SessionModel, error) {
	records, err := session.db.ReadRecords(&session)
	if err != nil {
		return nil, err
	}

	sessions := make([]*SessionModel, 0)
	for _, record := range records {
		session := NewSessionModel(session.db)
		err = session.Scan(record)
		if err != nil {
			return nil, err
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (session *SessionModel) Read() error {
	return session.db.ReadRecord(session)
}

func (session *SessionModel) Update() error {
	return session.db.UpdateRecord(session)
}

func (session *SessionModel) AddSubmission(submission SubmissionData) error {
	submissionCount := 0
	for _, existingSubmission := range session.Data.Submissions {
		if existingSubmission.UserId == submission.UserId {
			submissionCount++
		}
	}

	if submissionCount >= session.Data.MaxSubmissions {
		return ErrSubmissionsMaxedOut
	}

	session.Data.Submissions = append(session.Data.Submissions, submission)

	return nil
}

func (session *SessionModel) GetSubmission(submissionId string) (*SubmissionData, error) {
	for _, submission := range session.Data.Submissions {
		if submission.Id == submissionId {
			return &submission, nil
		}
	}

	return nil, ErrSubmissionNotFound
}

func (session *SessionModel) DeleteSubmission(submissionId string, userId int64) error {
	for i, submission := range session.Data.Submissions {
		if submission.Id == submissionId {
			if submission.UserId != userId {
				return ErrSubmissionUpdateUnauthorized
			}

			session.Data.Submissions = append(session.Data.Submissions[:i], session.Data.Submissions[i+1:]...)
			return nil
		}
	}

	return ErrSubmissionNotFound
}

func (session *SessionModel) AddVote(vote VoteData) {
	session.Data.Votes = append(session.Data.Votes, vote)
}

func (session *SessionModel) GetVote(voteId string, userId int64) (*VoteData, error) {
	for _, vote := range session.Data.Votes {
		if vote.Id == voteId {
			if vote.UserId != userId {
				return nil, ErrVoteUpdateUnauthorized
			}

			return &vote, nil
		}
	}

	return nil, ErrVoteNotFound
}

func (session *SessionModel) DeleteVote(voteId string, userId int64) error {
	for i, vote := range session.Data.Votes {
		if vote.Id == voteId {
			if vote.UserId != userId {
				return ErrVoteUpdateUnauthorized
			}

			session.Data.Votes = append(session.Data.Votes[:i], session.Data.Votes[i+1:]...)
			return nil
		}
	}

	return ErrVoteNotFound
}

func (session *SessionModel) Phase() SessionPhase {
	if time.Since(session.Data.StartAt) < session.Data.SubmissionDuration {
		return SubmissionPhase
	}

	if time.Since(session.Data.StartAt) < session.Data.SubmissionDuration+session.Data.VoteDuration {
		return VotePhase
	}

	return ResultPhase
}

func (session *SessionModel) PhaseDurationRemaining() time.Duration {
	sessionPhase := session.Phase()

	switch {
	case sessionPhase == SubmissionPhase:
		return session.Data.SubmissionDuration - time.Since(session.Data.StartAt)
	case sessionPhase == VotePhase:
		return session.Data.SubmissionDuration + session.Data.VoteDuration - time.Since(session.Data.StartAt)
	default:
		return 0
	}
}

func (session *SessionModel) PlaylistName() string {
	return fmt.Sprintf("%s %s", session.Data.Name, session.Data.StartAt.Format("2006-01-02"))
}
