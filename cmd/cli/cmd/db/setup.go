package db

import (
	"log"

	"github.com/CaribouBlue/mixtape/internal/storage"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup a SQLite database",
	Run: func(cmd *cobra.Command, args []string) {
		dbPath := flagDbPath

		log.Println("Setting up the database @", dbPath)

		db, err := storage.NewSqliteDb(dbPath)
		if err != nil {
			log.Fatalln("Failed to connect to the database:", err)
			return
		}
		defer db.Close()

		createTables(db)
		createViews(db)

		log.Println("Database setup completed successfully.")
	},
}

func init() {
}

func createTables(db *storage.SqliteStore) {
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
		`CREATE TABLE IF NOT EXISTS ` + storage.TableNamePlayers + ` (
			session_id INTEGER,
			player_id INTEGER,
			playlist_id TEXT,
			FOREIGN KEY (session_id) REFERENCES ` + storage.TableNameSessions + ` (id),
			FOREIGN KEY (player_id) REFERENCES ` + storage.TableNameUsers + ` (id),
			PRIMARY KEY (session_id, player_id)
		);`,
		`CREATE TABLE IF NOT EXISTS ` + storage.TableNameCandidates + ` (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			nominator_id INTEGER,
			session_id INTEGER,
			track_id TEXT,
			FOREIGN KEY (nominator_id) REFERENCES ` + storage.TableNameUsers + ` (id),
			FOREIGN KEY (session_id) REFERENCES ` + storage.TableNameSessions + ` (id)
		);`,
		`CREATE TABLE IF NOT EXISTS ` + storage.TableNameVotes + ` (
			session_id INTEGER,
			voter_id INTEGER,
			candidate_id INTEGER,
			FOREIGN KEY (session_id) REFERENCES ` + storage.TableNameSessions + ` (id),
			FOREIGN KEY (voter_id) REFERENCES ` + storage.TableNameUsers + ` (id),
			FOREIGN KEY (candidate_id) REFERENCES ` + storage.TableNameCandidates + ` (id),
			PRIMARY KEY (session_id, voter_id, candidate_id)
		);`,
	}

	for _, query := range tables {
		if _, err := db.Exec(query); err != nil {
			log.Fatalln("Failed to create table:", err)
		}
	}
}

func createViews(db *storage.SqliteStore) {
	views := []string{}

	for _, query := range views {
		if _, err := db.Exec(query); err != nil {
			log.Fatalln("Failed to create view:", err)
		}
	}
}
