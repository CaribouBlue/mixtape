package db

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
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
	Id          int64                 `json:"id"`
	Submissions []SubmissionDataModel `json:"submissions"`
	Votes       []VoteDataModel       `json:"votes"`
	CreatedAt   time.Time             `json:"createdAt"`
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

func NewGameSessionDataModel() *GameSessionDataModel {
	return &GameSessionDataModel{
		Submissions: make([]SubmissionDataModel, 0),
		Votes:       make([]VoteDataModel, 0),
		CreatedAt:   time.Now(),
	}
}
