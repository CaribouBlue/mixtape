package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSubmissionsMaxedOut          = errors.New("user has already submitted the maximum number of submissions")
	ErrSubmissionNotFound           = errors.New("no submission found with the given ID")
	ErrSubmissionUpdateUnauthorized = errors.New("user is unauthorized to update submission")
)

type SubmissionDataModel struct {
	Id      string `json:"id"`
	UserId  int64  `json:"userId"`
	TrackId string `json:"trackId"`
}

func NewSubmissionDataModel(userId int64, trackId string) *SubmissionDataModel {
	return &SubmissionDataModel{
		Id:      uuid.New().String(),
		UserId:  userId,
		TrackId: trackId,
	}
}

type VoteDataModel struct {
	Id           string `json:"id"`
	UserId       int64  `json:"userId"`
	SubmissionId int64  `json:"submissionId"`
}

func NewVoteDataModel(userId, submissionId int64) *VoteDataModel {
	return &VoteDataModel{
		Id:           uuid.New().String(),
		UserId:       userId,
		SubmissionId: submissionId,
	}
}

type GameSessionDataModel struct {
	Id                 int64                 `json:"id"`
	Submissions        []SubmissionDataModel `json:"submissions"`
	Votes              []VoteDataModel       `json:"votes"`
	CreatedAt          time.Time             `json:"createdAt"`
	MaxSubmissions     int                   `json:"maxSubmissions"`
	StartAt            time.Time             `json:"startAt"`
	SubmissionDuration time.Duration         `json:"submissionDuration"`
	VoteDuration       time.Duration         `json:"voteDuration"`
}

func (gameSession *GameSessionDataModel) GetTableName() string {
	return "game_session"
}

func (gameSession *GameSessionDataModel) SetId(id int64) {
	gameSession.Id = id
}

func (gameSession *GameSessionDataModel) GetId() int64 {
	return gameSession.Id
}

func (gameSession *GameSessionDataModel) Scan(value interface{}) error {
	return json.Unmarshal([]byte(value.(string)), gameSession)
}

func (gameSession *GameSessionDataModel) Value() (driver.Value, error) {
	return json.Marshal(gameSession)
}

func (gameSession *GameSessionDataModel) Insert() error {
	_, err := insertJsonDataModel(gameSession)
	return err
}

func (gameSession *GameSessionDataModel) Update() error {
	return updateJsonDataModel(gameSession)
}

func (gameSession *GameSessionDataModel) GetById() error {
	_, err := getJsonDataModelById(gameSession)
	return err
}

func (gameSession *GameSessionDataModel) AddSubmission(submission SubmissionDataModel) error {
	submissionCount := 0
	for _, existingSubmission := range gameSession.Submissions {
		if existingSubmission.UserId == submission.UserId {
			submissionCount++
		}
	}

	if submissionCount >= gameSession.MaxSubmissions {
		return ErrSubmissionsMaxedOut
	}

	gameSession.Submissions = append(gameSession.Submissions, submission)

	return nil
}

func (gameSession *GameSessionDataModel) DeleteSubmission(submissionId string, userId int64) error {
	for i, submission := range gameSession.Submissions {
		if submission.Id == submissionId {
			if submission.UserId != userId {
				return ErrSubmissionUpdateUnauthorized
			}

			gameSession.Submissions = append(gameSession.Submissions[:i], gameSession.Submissions[i+1:]...)
			return nil
		}
	}

	return ErrSubmissionNotFound
}

func (gameSession *GameSessionDataModel) GetSubmission(submissionId string, userId int64) (*SubmissionDataModel, error) {
	for _, submission := range gameSession.Submissions {
		if submission.Id == submissionId {
			if submission.UserId != userId {
				return nil, ErrSubmissionUpdateUnauthorized
			}

			return &submission, nil
		}
	}

	return nil, ErrSubmissionNotFound
}

func (gameSession *GameSessionDataModel) SubmissionDaysLeft() int {
	var daysLeft int
	submissionDurationLeft := gameSession.SubmissionDuration - time.Since(gameSession.CreatedAt)
	if submissionDurationLeft < 0 {
		daysLeft = 0
	} else {
		daysLeft = int(submissionDurationLeft.Hours() / 24)
	}

	return daysLeft
}

func NewGameSessionDataModel() *GameSessionDataModel {
	dayDuration := time.Hour * 24
	return &GameSessionDataModel{
		Submissions:        make([]SubmissionDataModel, 0),
		Votes:              make([]VoteDataModel, 0),
		CreatedAt:          time.Now(),
		MaxSubmissions:     5,
		SubmissionDuration: 5 * dayDuration,
		VoteDuration:       14 * dayDuration,
	}
}
