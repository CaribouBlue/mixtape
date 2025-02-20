package main

import (
	"log"
	"os"
	"time"

	"github.com/CaribouBlue/top-spot/internal/core"
	"github.com/CaribouBlue/top-spot/internal/storage"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	log.Println("Setting up the database...")

	dbPath := os.Getenv("DB_PATH")
	e := os.Remove(dbPath)
	if e != nil && !os.IsNotExist(e) {
		log.Fatal(e)
	}

	db, err := storage.NewSqliteDb(dbPath)
	if err != nil {
		log.Fatalln("Failed to connect to the database:", err)
		return
	}
	defer db.Close()

	CreateTables(db)
	CreateViews(db)

	log.Println("Database setup completed successfully.")
	log.Println("Adding test data...")

	log.Default().Println("Creating users...")
	CreateUsers(db)

	log.Default().Println("Creating submission phase session...")
	CreateSubmissionPhaseSession(db)

	log.Default().Println("Creating vote phase session...")
	CreateVotePhaseSession(db)

	log.Default().Println("Creating result phase session...")
	CreateResultPhaseSession(db)

	log.Println("Test data added successfully.")
}

func CreateTables(db *storage.SqliteStore) {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS ` + storage.TableNameUsers + ` (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT,
			display_name TEXT,
			hashed_password TEXT,
			spotify_token TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS ` + storage.TableNameSessions + ` (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			created_by INTEGER,
			created_at INTEGER,
			max_submissions INTEGER,
			start_at INTEGER,
			submission_phase_duration INTEGER,
			vote_phase_duration INTEGER,
			FOREIGN KEY (created_by) REFERENCES ` + storage.TableNameUsers + ` (id)
		);`,
		`CREATE TABLE IF NOT EXISTS ` + storage.TableNameCandidates + ` (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			session_id INTEGER,
			track_id TEXT,
			FOREIGN KEY (user_id) REFERENCES ` + storage.TableNameUsers + ` (id),
			FOREIGN KEY (session_id) REFERENCES ` + storage.TableNameSessions + ` (id)
		);`,
		`CREATE TABLE IF NOT EXISTS ` + storage.TableNameVotes + ` (
			session_id INTEGER,
			user_id INTEGER,
			candidate_id INTEGER,
			FOREIGN KEY (session_id) REFERENCES ` + storage.TableNameSessions + ` (id),
			FOREIGN KEY (user_id) REFERENCES ` + storage.TableNameUsers + ` (id),
			FOREIGN KEY (candidate_id) REFERENCES ` + storage.TableNameCandidates + ` (id)
		);`,
		`CREATE TABLE IF NOT EXISTS ` + storage.TableNamePlaylists + ` (
			session_id INTEGER,
			user_id INTEGER,
			playlist_id TEXT,
			FOREIGN KEY (session_id) REFERENCES ` + storage.TableNameSessions + ` (id),
			FOREIGN KEY (user_id) REFERENCES ` + storage.TableNameUsers + ` (id),
			PRIMARY KEY (session_id, user_id)
		);`,
	}

	for _, query := range tables {
		if _, err := db.Exec(query); err != nil {
			log.Fatalln("Failed to create table:", err)
		}
	}
}

func CreateViews(db *storage.SqliteStore) {
	views := []string{}

	for _, query := range views {
		if _, err := db.Exec(query); err != nil {
			log.Fatalln("Failed to create view:", err)
		}
	}
}

func CreateUsers(db *storage.SqliteStore) {
	defaultHashedPassword, err := core.HashPassword("pwd")
	if err != nil {
		log.Fatalln("Error hashing password:", err)
	}

	_, err = db.CreateUser(&core.UserEntity{
		Username:       "alice",
		DisplayName:    "alice",
		HashedPassword: defaultHashedPassword,
		SpotifyToken:   "",
	})
	if err != nil {
		log.Fatalln("Error creating user alice:", err)
	}

	_, err = db.CreateUser(&core.UserEntity{
		Username:       "bob",
		DisplayName:    "bob",
		HashedPassword: defaultHashedPassword,
		SpotifyToken:   "",
	})
	if err != nil {
		log.Fatalln("Error creating user bob:", err)
	}

	_, err = db.CreateUser(&core.UserEntity{
		Username:       "john",
		DisplayName:    "john",
		HashedPassword: defaultHashedPassword,
		SpotifyToken:   "",
	})
	if err != nil {
		log.Fatalln("Error creating user john:", err)
	}

	_, err = db.CreateUser(&core.UserEntity{
		Username:       "jane",
		DisplayName:    "jane",
		HashedPassword: defaultHashedPassword,
		SpotifyToken:   "",
	})
	if err != nil {
		log.Fatalln("Error creating user jane:", err)
	}
}

func CreateSubmissionPhaseSession(db *storage.SqliteStore) {
	session := core.NewSessionEntity("Submission Phase Session", 1)
	_, err := db.CreateSession(session)
	if err != nil {
		log.Fatalln("Error creating session:", err)
	}
}

func CreateVotePhaseSession(db *storage.SqliteStore) {
	session := core.NewSessionEntity("Vote Phase Session", 1)
	session.StartAt = session.StartAt.Add(-24 * time.Hour)
	session.SubmissionPhaseDuration = 24 * time.Hour
	session.VotePhaseDuration = 24 * time.Hour
	_, err := db.CreateSession(session)
	if err != nil {
		log.Fatalln("Error creating session:", err)
	}

	candidates := []core.CandidateEntity{
		{
			UserId:    1,
			SessionId: session.Id,
			TrackId:   "6gH1UKDAhWS6qXzKXB4wuY",
		},
		{
			UserId:    1,
			SessionId: session.Id,
			TrackId:   "7qwt4xUIqQWCu1DJf96g2k",
		},
		{
			UserId:    1,
			SessionId: session.Id,
			TrackId:   "1rqduvolf1CVHSzY519bPp",
		},
		{
			UserId:    1,
			SessionId: session.Id,
			TrackId:   "3sl4dcqSwxHVnLfqwF2jly",
		},
		{
			UserId:    1,
			SessionId: session.Id,
			TrackId:   "62PaSfnXSMyLshYJrlTuL3",
		},
		{
			UserId:    2,
			SessionId: session.Id,
			TrackId:   "1j8xbu9phaY9wNAaUSAqVf",
		},
		{
			UserId:    2,
			SessionId: session.Id,
			TrackId:   "6quGF3Kvzd5WYEEuCmvCe1",
		},
		{
			UserId:    2,
			SessionId: session.Id,
			TrackId:   "3HGwI9qwq5XqBDeZBV3zti",
		},
		{
			UserId:    2,
			SessionId: session.Id,
			TrackId:   "5Y9HJkaDmUlIfgNZzUYd5x",
		},
		{
			UserId:    2,
			SessionId: session.Id,
			TrackId:   "3ApxpM5ghkdjWKhbrQaPLk",
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
	session := core.NewSessionEntity("Result Phase Session", 1)
	session.StartAt = session.StartAt.Add(-2 * time.Hour)
	session.SubmissionPhaseDuration = 1 * time.Hour
	session.VotePhaseDuration = 1 * time.Hour
	_, err := db.CreateSession(session)
	if err != nil {
		log.Fatalln("Error creating session:", err)
	}

	candidates := []core.CandidateEntity{
		{
			UserId:    1,
			SessionId: session.Id,
			TrackId:   "6gH1UKDAhWS6qXzKXB4wuY",
		},
		{
			UserId:    1,
			SessionId: session.Id,
			TrackId:   "7qwt4xUIqQWCu1DJf96g2k",
		},
		{
			UserId:    1,
			SessionId: session.Id,
			TrackId:   "1rqduvolf1CVHSzY519bPp",
		},
		{
			UserId:    1,
			SessionId: session.Id,
			TrackId:   "3sl4dcqSwxHVnLfqwF2jly",
		},
		{
			UserId:    1,
			SessionId: session.Id,
			TrackId:   "62PaSfnXSMyLshYJrlTuL3",
		},
		{
			UserId:    2,
			SessionId: session.Id,
			TrackId:   "1j8xbu9phaY9wNAaUSAqVf",
		},
		{
			UserId:    2,
			SessionId: session.Id,
			TrackId:   "6quGF3Kvzd5WYEEuCmvCe1",
		},
		{
			UserId:    2,
			SessionId: session.Id,
			TrackId:   "3HGwI9qwq5XqBDeZBV3zti",
		},
		{
			UserId:    2,
			SessionId: session.Id,
			TrackId:   "5Y9HJkaDmUlIfgNZzUYd5x",
		},
		{
			UserId:    2,
			SessionId: session.Id,
			TrackId:   "3ApxpM5ghkdjWKhbrQaPLk",
		},
	}

	for _, candidate := range candidates {
		_, err = db.AddCandidate(session.Id, &candidate)
		if err != nil {
			log.Fatalln("Error adding candidate for track ", candidate.TrackId, ": ", err)
		}
	}

}
