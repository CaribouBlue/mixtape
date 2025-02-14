package main

import (
	"log"
	"os"
	"path"
	"time"

	"github.com/CaribouBlue/top-spot/internal/appdata"
	"github.com/CaribouBlue/top-spot/internal/core"
	"github.com/CaribouBlue/top-spot/internal/storage"
)

func main() {
	log.Println("Setting up the database...")
	appDataDir, err := appdata.GetAppDataDir()
	if err != nil {
		log.Fatal(err)
	}

	dbPath := path.Join(appDataDir, "/top-spot.db")
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

	userService := core.NewUserService(db)
	sessionService := core.NewSessionService(db)

	CreateUsers(userService)
	CreateSubmissionPhaseSession(sessionService)
	CreateVotePhaseSession(sessionService)

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
		`CREATE TABLE IF NOT EXISTS ` + storage.TableNameSubmissions + ` (
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
			submission_id INTEGER,
			FOREIGN KEY (session_id) REFERENCES ` + storage.TableNameSessions + ` (id),
			FOREIGN KEY (user_id) REFERENCES ` + storage.TableNameUsers + ` (id),
			FOREIGN KEY (submission_id) REFERENCES ` + storage.TableNameSubmissions + ` (id)
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
	views := []string{
		`CREATE VIEW IF NOT EXISTS ` + storage.ViewNameCandidates + ` AS
			SELECT s.id AS submission_id,
				s.user_id AS submission_user_id,
				s.session_id AS session_id,
				s.track_id AS track_id,
				v.submission_id AS vote_submission_id,
				u.id AS vote_user_id
			FROM ` + storage.TableNameSubmissions + ` s
				FULL JOIN ` + storage.TableNameUsers + ` u ON u.id != s.user_id
				FULL JOIN ` + storage.TableNameVotes + ` v ON s.id = v.submission_id AND u.id = v.user_id;
`,
	}

	for _, query := range views {
		if _, err := db.Exec(query); err != nil {
			log.Fatalln("Failed to create view:", err)
		}
	}
}

func CreateUsers(userService *core.UserService) {
	_, err := userService.SignUpNewUser("BoB", "pwd", "pwd")
	if err != nil && err != core.ErrUsernameAlreadyExists {
		log.Fatalln("Error creating user BoB:", err)
	}

	_, err = userService.SignUpNewUser("Alice", "pwd", "pwd")
	if err != nil && err != core.ErrUsernameAlreadyExists {
		log.Fatalln("Error creating user Alice:", err)
	}

	_, err = userService.SignUpNewUser("John", "pwd", "pwd")
	if err != nil && err != core.ErrUsernameAlreadyExists {
		log.Fatalln("Error creating user John:", err)
	}

	_, err = userService.SignUpNewUser("J A N E", "pwd", "pwd")
	if err != nil && err != core.ErrUsernameAlreadyExists {
		log.Fatalln("Error creating user Jane:", err)
	}
}

func CreateSessions(sessionService *core.SessionService) {
	_, err := sessionService.CreateSession(core.NewSessionEntity("Test Session 1", 1))
	if err != nil {
		log.Fatalln("Error creating Test Session 1:", err)
	}

	_, err = sessionService.CreateSession(core.NewSessionEntity("Test Session 2", 1))
	if err != nil {
		log.Fatalln("Error creating Test Session 2:", err)
	}

	_, err = sessionService.CreateSession(core.NewSessionEntity("Test Session 3", 1))
	if err != nil {
		log.Fatalln("Error creating Test Session 3:", err)
	}
}

func CreateSubmissionPhaseSession(sessionService *core.SessionService) {
	_, err := sessionService.CreateSession(core.NewSessionEntity("Submission Phase Session", 1))
	if err != nil {
		log.Fatalln("Error creating Submission Phase Session:", err)
	}
}

func CreateVotePhaseSession(sessionService *core.SessionService) {
	session, err := sessionService.CreateSession(core.NewSessionEntity("Vote Phase Session", 1, core.WithSessionStartAt(time.Now().Add(-time.Hour*24)), core.WithSubmissionDuration(time.Hour*24)))
	if err != nil {
		log.Fatalln("Error creating Vote Phase Session:", err)
	}

	_, err = sessionService.AddUserSubmission(session.Id, 1, "track_id_1")
	if err != nil {
		log.Fatalln("Error creating submission 1:", err, "\nFor session:", session.Id)
	}

	_, err = sessionService.AddUserSubmission(session.Id, 1, "track_id_2")
	if err != nil {
		log.Fatalln("Error creating submission 2:", err, "\nFor session:", session.Id)
	}

	_, err = sessionService.AddUserSubmission(session.Id, 1, "track_id_3")
	if err != nil {
		log.Fatalln("Error creating submission 3:", err, "\nFor session:", session.Id)
	}

	_, err = sessionService.AddUserSubmission(session.Id, 1, "track_id_4")
	if err != nil {
		log.Fatalln("Error creating submission 4:", err, "\nFor session:", session.Id)
	}

	_, err = sessionService.AddUserSubmission(session.Id, 1, "track_id_5")
	if err != nil {
		log.Fatalln("Error creating submission 5:", err, "\nFor session:", session.Id)
	}

	_, err = sessionService.AddUserSubmission(session.Id, 2, "track_id_6")
	if err != nil {
		log.Fatalln("Error creating submission 6:", err, "\nFor session:", session.Id)
	}

	_, err = sessionService.AddUserSubmission(session.Id, 2, "track_id_7")
	if err != nil {
		log.Fatalln("Error creating submission 7:", err, "\nFor session:", session.Id)
	}

	_, err = sessionService.AddUserSubmission(session.Id, 2, "track_id_8")
	if err != nil {
		log.Fatalln("Error creating submission 8:", err, "\nFor session:", session.Id)
	}

	_, err = sessionService.AddUserSubmission(session.Id, 2, "track_id_9")
	if err != nil {
		log.Fatalln("Error creating submission 9:", err, "\nFor session:", session.Id)
	}

	_, err = sessionService.AddUserSubmission(session.Id, 2, "track_id_10")
	if err != nil {
		log.Fatalln("Error creating submission 10:", err, "\nFor session:", session.Id)
	}
}

func CreateSubmissions(sessionService *core.SessionService) {
	_, err := sessionService.AddUserSubmission(1, 1, "track_id_1")
	if err != nil {
		log.Fatalln("Error creating submission 1:", err)
	}

	_, err = sessionService.AddUserSubmission(1, 1, "track_id_2")
	if err != nil {
		log.Fatalln("Error creating submission 2:", err)
	}

	_, err = sessionService.AddUserSubmission(1, 1, "track_id_3")
	if err != nil {
		log.Fatalln("Error creating submission 3:", err)
	}

	_, err = sessionService.AddUserSubmission(1, 1, "track_id_4")
	if err != nil {
		log.Fatalln("Error creating submission 4:", err)
	}

	_, err = sessionService.AddUserSubmission(1, 1, "track_id_5")
	if err != nil {
		log.Fatalln("Error creating submission 5:", err)
	}

	_, err = sessionService.AddUserSubmission(1, 2, "track_id_6")
	if err != nil {
		log.Fatalln("Error creating submission 6:", err)
	}

	_, err = sessionService.AddUserSubmission(1, 2, "track_id_7")
	if err != nil {
		log.Fatalln("Error creating submission 7:", err)
	}

	_, err = sessionService.AddUserSubmission(1, 2, "track_id_8")
	if err != nil {
		log.Fatalln("Error creating submission 8:", err)
	}

	_, err = sessionService.AddUserSubmission(1, 2, "track_id_9")
	if err != nil {
		log.Fatalln("Error creating submission 9:", err)
	}

	_, err = sessionService.AddUserSubmission(1, 2, "track_id_10")
	if err != nil {
		log.Fatalln("Error creating submission 10:", err)
	}
}

func CreateVotes(sessionService *core.SessionService) {
	_, err := sessionService.VoteForCandidate(1, 2, 1)
	if err != nil {
		log.Fatalln("Error creating vote 1:", err)
	}

	_, err = sessionService.VoteForCandidate(1, 2, 2)
	if err != nil {
		log.Fatalln("Error creating vote 2:", err)
	}

	_, err = sessionService.VoteForCandidate(1, 1, 6)
	if err != nil {
		log.Fatalln("Error creating vote 3:", err)
	}

	_, err = sessionService.VoteForCandidate(1, 1, 7)
	if err != nil {
		log.Fatalln("Error creating vote 4:", err)
	}

	_, err = sessionService.VoteForCandidate(1, 3, 1)
	if err != nil {
		log.Fatalln("Error creating vote 5:", err)
	}

	_, err = sessionService.VoteForCandidate(1, 3, 2)
	if err != nil {
		log.Fatalln("Error creating vote 6:", err)
	}

	_, err = sessionService.VoteForCandidate(1, 4, 10)
	if err != nil {
		log.Fatalln("Error creating vote 7:", err)
	}

}
