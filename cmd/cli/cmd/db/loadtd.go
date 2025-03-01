package db

import (
	"log"
	"time"

	"github.com/CaribouBlue/mixtape/internal/core"
	"github.com/CaribouBlue/mixtape/internal/storage"
	"github.com/spf13/cobra"
)

var loadTestDataCmd = &cobra.Command{
	Use:   "loadtd",
	Short: "Load test data into a SQLite database",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dbPath := flagDbPath

		db, err := storage.NewSqliteDb(dbPath)
		if err != nil {
			log.Fatalln("Failed to connect to the database:", err)
			return
		}
		defer db.Close()

		log.Println("Adding test data to the database @", dbPath)

		log.Default().Println("Creating users...")
		CreateUsers(db)

		log.Default().Println("Creating submission phase session...")
		CreateSubmissionPhaseSession(db)

		log.Default().Println("Creating vote phase session...")
		CreateVotePhaseSession(db)

		log.Default().Println("Creating result phase session...")
		CreateResultPhaseSession(db)

		log.Println("Test data added successfully")
	},
}

var defaultHashedPassword, defaultHashedPasswordErr = core.HashPassword("pwd")

var mockUserAlice = &core.UserEntity{
	Username:       "alice",
	DisplayName:    "alice",
	HashedPassword: defaultHashedPassword,
	SpotifyToken:   "",
}

var mockUserBob = &core.UserEntity{
	Username:       "bob",
	DisplayName:    "bob",
	HashedPassword: defaultHashedPassword,
	SpotifyToken:   "",
}

var mockUserJohn = &core.UserEntity{
	Username:       "john",
	DisplayName:    "john",
	HashedPassword: defaultHashedPassword,
	SpotifyToken:   "",
}

var mockUserJane = &core.UserEntity{
	Username:       "jane",
	DisplayName:    "jane",
	HashedPassword: defaultHashedPassword,
	SpotifyToken:   "",
}

func init() {
	if defaultHashedPasswordErr != nil {
		log.Fatalln("Error hashing default password:", defaultHashedPasswordErr)
	}
}

func CreateUsers(db *storage.SqliteStore) {
	users := []*core.UserEntity{
		mockUserAlice,
		mockUserBob,
		mockUserJohn,
		mockUserJane,
	}

	for _, user := range users {
		newUser, err := db.CreateUser(user)
		if err != nil {
			log.Fatalf("Error creating user %s: %v", user.Username, err)
		}
		*user = *newUser
	}
}

func CreateSubmissionPhaseSession(db *storage.SqliteStore) {
	session := core.NewSessionEntity("Submission Phase Session", mockUserAlice.Id)
	_, err := db.CreateSession(session)
	if err != nil {
		log.Fatalln("Error creating session:", err)
	}
	_, err = db.AddPlayer(session.Id, &core.PlayerEntity{
		SessionId: session.Id,
		PlayerId:  mockUserAlice.Id,
	})
	if err != nil {
		log.Fatalln("Error adding player:", err)
	}
}

func CreateVotePhaseSession(db *storage.SqliteStore) {
	session := core.NewSessionEntity("Vote Phase Session", mockUserAlice.Id)
	session.StartAt = session.StartAt.Add(-24 * time.Hour)
	session.SubmissionPhaseDuration = 24 * time.Hour
	session.VotePhaseDuration = 24 * time.Hour
	_, err := db.CreateSession(session)
	if err != nil {
		log.Fatalln("Error creating session:", err)
	}

	_, err = db.AddPlayer(session.Id, &core.PlayerEntity{
		SessionId: session.Id,
		PlayerId:  mockUserAlice.Id,
	})
	if err != nil {
		log.Fatalf("Error creating player %s: %v", mockUserAlice.Username, err)
	}

	_, err = db.AddPlayer(session.Id, &core.PlayerEntity{
		SessionId: session.Id,
		PlayerId:  mockUserBob.Id,
	})
	if err != nil {
		log.Fatalf("Error creating player %s: %v", mockUserBob.Username, err)
	}

	candidates := []core.CandidateEntity{
		{
			NominatorId: mockUserAlice.Id,
			SessionId:   session.Id,
			TrackId:     "6gH1UKDAhWS6qXzKXB4wuY",
		},
		{
			NominatorId: mockUserAlice.Id,
			SessionId:   session.Id,
			TrackId:     "7qwt4xUIqQWCu1DJf96g2k",
		},
		{
			NominatorId: mockUserAlice.Id,
			SessionId:   session.Id,
			TrackId:     "1rqduvolf1CVHSzY519bPp",
		},
		{
			NominatorId: mockUserAlice.Id,
			SessionId:   session.Id,
			TrackId:     "3sl4dcqSwxHVnLfqwF2jly",
		},
		{
			NominatorId: mockUserAlice.Id,
			SessionId:   session.Id,
			TrackId:     "62PaSfnXSMyLshYJrlTuL3",
		},
		{
			NominatorId: mockUserBob.Id,
			SessionId:   session.Id,
			TrackId:     "1j8xbu9phaY9wNAaUSAqVf",
		},
		{
			NominatorId: mockUserBob.Id,
			SessionId:   session.Id,
			TrackId:     "6quGF3Kvzd5WYEEuCmvCe1",
		},
		{
			NominatorId: mockUserBob.Id,
			SessionId:   session.Id,
			TrackId:     "3HGwI9qwq5XqBDeZBV3zti",
		},
		{
			NominatorId: mockUserBob.Id,
			SessionId:   session.Id,
			TrackId:     "5Y9HJkaDmUlIfgNZzUYd5x",
		},
		{
			NominatorId: mockUserBob.Id,
			SessionId:   session.Id,
			TrackId:     "3ApxpM5ghkdjWKhbrQaPLk",
		},
	}
	for _, candidate := range candidates {
		_, err = db.AddCandidate(session.Id, &candidate)
		if err != nil {
			log.Fatalln("Error adding candidate for track ", candidate.TrackId, ": ", err)
		}
	}

}

func CreateResultPhaseSession(db *storage.SqliteStore) {
	session := core.NewSessionEntity("Result Phase Session", mockUserAlice.Id)
	session.StartAt = session.StartAt.Add(-2 * time.Hour)
	session.SubmissionPhaseDuration = 1 * time.Hour
	session.VotePhaseDuration = 1 * time.Hour
	_, err := db.CreateSession(session)
	if err != nil {
		log.Fatalln("Error creating session:", err)
	}

	_, err = db.AddPlayer(session.Id, &core.PlayerEntity{
		SessionId: session.Id,
		PlayerId:  mockUserAlice.Id,
	})
	if err != nil {
		log.Fatalf("Error creating player %s: %v", mockUserAlice.Username, err)
	}

	_, err = db.AddPlayer(session.Id, &core.PlayerEntity{
		SessionId: session.Id,
		PlayerId:  mockUserBob.Id,
	})
	if err != nil {
		log.Fatalf("Error creating player %s: %v", mockUserBob.Username, err)
	}

	candidates := []core.CandidateEntity{
		{
			NominatorId: mockUserAlice.Id,
			SessionId:   session.Id,
			TrackId:     "6gH1UKDAhWS6qXzKXB4wuY",
		},
		{
			NominatorId: mockUserAlice.Id,
			SessionId:   session.Id,
			TrackId:     "7qwt4xUIqQWCu1DJf96g2k",
		},
		{
			NominatorId: mockUserAlice.Id,
			SessionId:   session.Id,
			TrackId:     "1rqduvolf1CVHSzY519bPp",
		},
		{
			NominatorId: mockUserAlice.Id,
			SessionId:   session.Id,
			TrackId:     "3sl4dcqSwxHVnLfqwF2jly",
		},
		{
			NominatorId: mockUserAlice.Id,
			SessionId:   session.Id,
			TrackId:     "62PaSfnXSMyLshYJrlTuL3",
		},
		{
			NominatorId: mockUserBob.Id,
			SessionId:   session.Id,
			TrackId:     "1j8xbu9phaY9wNAaUSAqVf",
		},
		{
			NominatorId: mockUserBob.Id,
			SessionId:   session.Id,
			TrackId:     "6quGF3Kvzd5WYEEuCmvCe1",
		},
		{
			NominatorId: mockUserBob.Id,
			SessionId:   session.Id,
			TrackId:     "3HGwI9qwq5XqBDeZBV3zti",
		},
		{
			NominatorId: mockUserBob.Id,
			SessionId:   session.Id,
			TrackId:     "5Y9HJkaDmUlIfgNZzUYd5x",
		},
		{
			NominatorId: mockUserBob.Id,
			SessionId:   session.Id,
			TrackId:     "3ApxpM5ghkdjWKhbrQaPLk",
		},
	}

	for _, candidate := range candidates {
		_, err = db.AddCandidate(session.Id, &candidate)
		if err != nil {
			log.Fatalln("Error adding candidate for track ", candidate.TrackId, ": ", err)
		}
	}

}
