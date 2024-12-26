package main

import (
	"log"
	"os"
	"path"

	"github.com/CaribouBlue/top-spot/internal/appdata"
	"github.com/CaribouBlue/top-spot/internal/db"
	"github.com/CaribouBlue/top-spot/internal/model"
)

var MockSubmissions = []model.SubmissionData{
	{
		Id:      "f9b7dc17-081f-428f-80ff-b27db0bbe5f5",
		UserId:  6666,
		TrackId: "6gH1UKDAhWS6qXzKXB4wuY",
	},
	{
		Id:      "e0b6ed66-c995-4519-ad18-39acdfdd5739",
		UserId:  6666,
		TrackId: "7qwt4xUIqQWCu1DJf96g2k",
	},
	{
		Id:      "93f886a6-effc-4239-910f-9e004a9b69f5",
		UserId:  6666,
		TrackId: "1rqduvolf1CVHSzY519bPp",
	},
	{
		Id:      "69f48d20-3640-4700-813f-6697753223c0",
		UserId:  6666,
		TrackId: "3sl4dcqSwxHVnLfqwF2jly",
	},
	{
		Id:      "06603499-1e7f-46f3-9f50-af17f27845ff",
		UserId:  6666,
		TrackId: "62PaSfnXSMyLshYJrlTuL3",
	},
	{
		Id:      "88fc2a5e-3e54-44c8-a8f5-9bce8e43b9e1",
		UserId:  9999,
		TrackId: "1j8xbu9phaY9wNAaUSAqVf",
	},
	{
		Id:      "e4b82abe-95c7-4e72-a21b-77c5bd222e3e",
		UserId:  9999,
		TrackId: "6quGF3Kvzd5WYEEuCmvCe1",
	},
	{
		Id:      "84b4d9a2-fbdc-4db6-b666-c0dbe377bc8e",
		UserId:  9999,
		TrackId: "3HGwI9qwq5XqBDeZBV3zti",
	},
	{
		Id:      "2842e488-79f4-4d55-954e-2e695c4cb036",
		UserId:  9999,
		TrackId: "5Y9HJkaDmUlIfgNZzUYd5x",
	},
	{
		Id:      "a7d19ca8-3a81-46c8-970f-4956ddc77028",
		UserId:  9999,
		TrackId: "3ApxpM5ghkdjWKhbrQaPLk",
	},
}

func main() {
	appDataDir, err := appdata.GetAppDataDir()
	if err != nil {
		log.Fatal(err)
	}

	dbPath := path.Join(appDataDir, db.GlobalDbName)
	e := os.Remove(dbPath)
	if e != nil && !os.IsNotExist(e) {
		log.Fatal(e)
	}

	database := db.NewSqliteJsonDb(db.GlobalDbName)

	for _, collection := range []db.Model{
		&model.SessionModel{},
		&model.UserModel{},
	} {
		err = database.NewCollection(collection)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = database.NewCollection(&model.SessionModel{})
	if err != nil {
		log.Fatal(err)
	}

	err = database.NewCollection(&model.UserModel{})
	if err != nil {
		log.Fatal(err)
	}

	err = addTestSessions(database)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Test data setup successfully")
}

func addTestSessions(database db.Db) error {
	newSession := model.NewSessionModel(database, model.WithId(0))
	newSession.Data.Name = "New Session"
	err := newSession.Create()
	if err != nil {
		return err
	}

	voteSession := model.NewSessionModel(database, model.WithId(1))
	voteSession.Data.Name = "Vote Session"
	voteSession.Data.Submissions = MockSubmissions
	voteSession.Data.StartAt = voteSession.Data.StartAt.Add(-voteSession.Data.SubmissionDuration)
	err = voteSession.Create()
	if err != nil {
		return err
	}

	resultsSession := model.NewSessionModel(database, model.WithId(2))
	resultsSession.Data.Name = "Results Session"
	resultsSession.Data.Submissions = MockSubmissions
	resultsSession.Data.StartAt = resultsSession.Data.StartAt.Add(-resultsSession.Data.SubmissionDuration - resultsSession.Data.VoteDuration)
	err = resultsSession.Create()
	if err != nil {
		return err
	}

	return nil
}
