package main

import (
	"log"
	"os"
	"path"

	"github.com/CaribouBlue/top-spot/internal/appdata"
	"github.com/CaribouBlue/top-spot/internal/db"
	"github.com/CaribouBlue/top-spot/internal/entities/session"
	"github.com/CaribouBlue/top-spot/internal/entities/user"
)

var MockSubmissions = []session.Submission{
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

var MockVotes = []session.Vote{
	{
		Id:           "0",
		UserId:       6666,
		SubmissionId: "f9b7dc17-081f-428f-80ff-b27db0bbe5f5",
	},
	{
		Id:           "1",
		UserId:       6666,
		SubmissionId: "a7d19ca8-3a81-46c8-970f-4956ddc77028",
	},
	{
		Id:           "1",
		UserId:       6666,
		SubmissionId: "84b4d9a2-fbdc-4db6-b666-c0dbe377bc8e",
	},
	{
		Id:           "2",
		UserId:       9999,
		SubmissionId: "84b4d9a2-fbdc-4db6-b666-c0dbe377bc8e",
	},
}

func main() {
	appDataDir, err := appdata.GetAppDataDir()
	if err != nil {
		log.Fatal(err)
	}

	dbPath := path.Join(appDataDir, "/top-spot.db")
	e := os.Remove(dbPath)
	if e != nil && !os.IsNotExist(e) {
		log.Fatal(e)
	}

	database, err := db.NewSqliteJsonDb(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	err = database.NewCollection(database.UserTable)
	if err != nil {
		log.Fatal(err)
	}

	err = database.NewCollection(database.SessionTable)
	if err != nil {
		log.Fatal(err)
	}

	userService := user.NewUserService(database)
	sessionService := session.NewSessionService(database)

	err = addTestUsers(userService)
	if err != nil {
		log.Fatal(err)
	}

	err = addTestSessions(sessionService)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Test data setup successfully")
}

func addTestUsers(userService user.UserService) error {
	_, err := userService.SignUp("bob", "pwd")
	if err != nil {
		return err
	}

	_, err = userService.SignUp("alice", "pwd")
	if err != nil {
		return err
	}

	return nil
}

func addTestSessions(sessionService session.SessionService) error {
	s := session.NewSession("New Session")
	s.Id = 0
	err := sessionService.Create(s)
	if err != nil {
		return err
	}

	s = session.NewSession("Vote Session")
	s.Id = 1
	s.Submissions = MockSubmissions
	s.StartAt = s.StartAt.Add(-s.SubmissionDuration)
	err = sessionService.Create(s)
	if err != nil {
		return err
	}

	s = session.NewSession("Results Session")
	s.Id = 2
	s.Submissions = MockSubmissions
	s.Votes = MockVotes
	s.StartAt = s.StartAt.Add(-s.SubmissionDuration - s.VoteDuration)
	err = sessionService.Create(s)
	if err != nil {
		return err
	}

	return nil
}
