package main

import (
	"log"
	"os"

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
	db, err := storage.NewSqliteDb(dbPath)
	if err != nil {
		log.Fatalln("Failed to connect to the database:", err)
		return
	}
	defer db.Close()

	CreateTables(db)
	CreateViews(db)

	log.Println("Database setup completed successfully.")
}

func CreateTables(db *storage.SqliteStore) {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS ` + storage.TableNameUsers + ` (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT,
			display_name TEXT,
			hashed_password TEXT,
			spotify_token TEXT,
			is_admin INTEGER DEFAULT (0)
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
