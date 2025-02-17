package storage

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/CaribouBlue/top-spot/internal/core"
	_ "github.com/mattn/go-sqlite3"
)

const (
	TableNameUsers      = "users"
	TableNameSessions   = "sessions"
	TableNameCandidates = "candidates"
	TableNameVotes      = "votes"
	TableNamePlaylists  = "playlists"
)

func makeSelectCandidatesQuery(conditional string) string {
	selectCandidatesQuery := `
		SELECT id,
			c.session_id AS session_id,
			c.user_id AS user_id,
			track_id,
			COUNT(DISTINCT v.user_id) AS votes
		FROM ` + TableNameCandidates + ` c
		FULL JOIN
			` + TableNameVotes + ` v ON v.candidate_id = c.id
		` + conditional + `
		GROUP BY c.id
	`
	return selectCandidatesQuery
}

type SqliteStore struct {
	dbPath string
	db     *sql.DB
}

func NewSqliteDb(dbPath string) (*SqliteStore, error) {
	sqlite := &SqliteStore{
		dbPath: dbPath,
	}
	err := sqlite.init()
	return sqlite, err
}

func (store *SqliteStore) init() error {
	db, err := sql.Open("sqlite3", store.dbPath)
	if err != nil {
		return err
	}
	store.db = db

	if err = store.db.Ping(); err != nil {
		return err
	}

	return nil
}

func (store *SqliteStore) Close() error {
	if store.db != nil {
		return store.db.Close()
	}
	return nil
}

func (store *SqliteStore) Exec(query string, args ...any) (sql.Result, error) {
	return store.db.Exec(query, args...)
}

// ------------------------------------------------------------
// | User Repository Methods
// ------------------------------------------------------------

func (store *SqliteStore) CreateUser(user *core.UserEntity) (*core.UserEntity, error) {
	query := "INSERT INTO " + TableNameUsers + " (username, display_name, hashed_password) VALUES (?, ?, ?)"
	result, err := store.Exec(query, user.Username, user.DisplayName, user.HashedPassword)
	if err != nil {
		return nil, err
	}

	userId, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	user.Id = userId

	return user, nil
}

func (store *SqliteStore) GetUserById(userId int64) (*core.UserEntity, error) {
	user := &core.UserEntity{}
	query := "SELECT id, username, display_name, spotify_token FROM " + TableNameUsers + " WHERE id = ?"
	row := store.db.QueryRow(query, userId)
	var spotifyToken sql.NullString
	err := row.Scan(&user.Id, &user.Username, &user.DisplayName, &spotifyToken)
	if err == sql.ErrNoRows {
		return nil, nil // User not found
	} else if err != nil {
		return nil, err
	}

	if spotifyToken.Valid {
		user.SpotifyToken = spotifyToken.String
	}

	return user, nil
}

func (store *SqliteStore) GetUserByUsername(username string) (*core.UserEntity, error) {
	user := &core.UserEntity{}
	query := "SELECT id, username, display_name, hashed_password FROM " + TableNameUsers + " WHERE username = ?"
	row := store.db.QueryRow(query, username)
	err := row.Scan(&user.Id, &user.Username, &user.DisplayName, &user.HashedPassword)
	if err == sql.ErrNoRows {
		return nil, nil // User not found
	} else if err != nil {
		return nil, err
	}

	return user, nil
}

func (store *SqliteStore) GetAllUsers() (*[]core.UserEntity, error) {
	query := "SELECT id, username, display_name FROM " + TableNameUsers
	rows, err := store.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]core.UserEntity, 0)
	for rows.Next() {
		user := core.UserEntity{}
		err := rows.Scan(&user.Id, &user.Username, &user.DisplayName)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &users, nil
}

func (store *SqliteStore) UpdateUserSpotifyToken(userId int64, spotifyToken string) (*core.UserEntity, error) {
	query := "UPDATE " + TableNameUsers + " SET spotify_token = ? WHERE id = ?"
	_, err := store.Exec(query, spotifyToken, userId)
	if err != nil {
		return nil, err
	}

	user, err := store.GetUserById(userId)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// ------------------------------------------------------------
// | Session Repository Methods
// ------------------------------------------------------------

func (store *SqliteStore) CreateSession(session *core.SessionEntity) (*core.SessionEntity, error) {
	query := `INSERT INTO ` + TableNameSessions + `
		(name, created_by, created_at, max_submissions, start_at, submission_phase_duration, vote_phase_duration) 
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	result, err := store.Exec(query, session.Name, session.CreatedBy, session.CreatedAt.Unix(), session.MaxSubmissions, session.StartAt.Unix(), session.SubmissionPhaseDuration, session.VotePhaseDuration)
	if err != nil {
		return nil, err
	}

	sessionId, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	session.Id = sessionId

	return session, nil
}

func (store *SqliteStore) GetSessionById(id int64) (*core.SessionEntity, error) {
	session := &core.SessionEntity{}
	query := "SELECT id, name, created_by, created_at, max_submissions, start_at, submission_phase_duration, vote_phase_duration FROM " + TableNameSessions + " WHERE id = ?"
	row := store.db.QueryRow(query, id)
	var CreatedAt, StartAt int64
	err := row.Scan(&session.Id, &session.Name, &session.CreatedBy, &CreatedAt, &session.MaxSubmissions, &StartAt, &session.SubmissionPhaseDuration, &session.VotePhaseDuration)
	if err == sql.ErrNoRows {
		return nil, nil // Session not found
	} else if err != nil {
		return nil, err
	}

	session.CreatedAt = time.Unix(CreatedAt, 0)
	session.StartAt = time.Unix(StartAt, 0)

	return session, nil
}

func (store *SqliteStore) GetAllSessions() (*[]core.SessionEntity, error) {
	query := "SELECT id, name, created_by, created_at, max_submissions, start_at, submission_phase_duration, vote_phase_duration FROM " + TableNameSessions
	rows, err := store.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	sessions := make([]core.SessionEntity, 0)
	for rows.Next() {
		session := core.SessionEntity{}
		var CreatedAt, StartAt int64
		err := rows.Scan(&session.Id, &session.Name, &session.CreatedBy, &CreatedAt, &session.MaxSubmissions, &StartAt, &session.SubmissionPhaseDuration, &session.VotePhaseDuration)
		if err != nil {
			return nil, err
		}

		session.CreatedAt = time.Unix(CreatedAt, 0)
		session.StartAt = time.Unix(StartAt, 0)

		sessions = append(sessions, session)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &sessions, nil
}

func (store *SqliteStore) AddCandidate(sessionId int64, candidate *core.CandidateEntity) (*core.CandidateEntity, error) {
	query := "INSERT INTO " + TableNameCandidates + " (session_id, user_id, track_id) VALUES (?, ?, ?)"
	result, err := store.Exec(query, sessionId, candidate.UserId, candidate.TrackId)
	if err != nil {
		return nil, err
	}

	candidateId, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	candidate.Id = candidateId

	return candidate, nil
}

func (store *SqliteStore) GetAllCandidates(sessionId int64) (*[]core.CandidateEntity, error) {
	conditional := fmt.Sprintf("WHERE c.session_id = %d", sessionId)
	rows, err := store.db.Query(makeSelectCandidatesQuery(conditional))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	candidates := make([]core.CandidateEntity, 0)
	for rows.Next() {
		candidate := core.CandidateEntity{}
		err := rows.Scan(&candidate.Id, &candidate.SessionId, &candidate.UserId, &candidate.TrackId, &candidate.Votes)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, candidate)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &candidates, nil
}

func (store *SqliteStore) GetCandidateById(sessionId int64, candidateId int64) (*core.CandidateEntity, error) {
	conditional := fmt.Sprintf("WHERE c.session_id = %d AND c.id = %d", sessionId, candidateId)
	row := store.db.QueryRow(makeSelectCandidatesQuery(conditional))
	candidate := &core.CandidateEntity{}
	err := row.Scan(&candidate.Id, &candidate.SessionId, &candidate.UserId, &candidate.TrackId, &candidate.Votes)
	if err == sql.ErrNoRows {
		return nil, nil // candidate not found
	} else if err != nil {
		return nil, err
	}
	return candidate, nil
}

func (store *SqliteStore) GetCandidatesByUserId(sessionId int64, userId int64) (*[]core.CandidateEntity, error) {
	conditional := fmt.Sprintf("WHERE c.session_id = %d AND c.user_id = %d", sessionId, userId)
	rows, err := store.db.Query(makeSelectCandidatesQuery(conditional))
	if err != nil {
		log.Default().Println("Error querying candidates: ", err)
		return nil, err
	}
	defer rows.Close()

	candidates := make([]core.CandidateEntity, 0)
	for rows.Next() {
		candidate := core.CandidateEntity{}
		err := rows.Scan(&candidate.Id, &candidate.SessionId, &candidate.UserId, &candidate.TrackId, &candidate.Votes)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, candidate)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &candidates, nil
}

func (store *SqliteStore) GetCandidateByNotUserId(sessionId int64, userId int64) (*[]core.CandidateEntity, error) {
	conditional := fmt.Sprintf("WHERE c.session_id = %d AND c.user_id != %d", sessionId, userId)
	rows, err := store.db.Query(makeSelectCandidatesQuery(conditional))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	candidates := make([]core.CandidateEntity, 0)
	for rows.Next() {
		candidate := core.CandidateEntity{}
		err := rows.Scan(&candidate.Id, &candidate.SessionId, &candidate.UserId, &candidate.TrackId, &candidate.Votes)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, candidate)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &candidates, nil
}

func (store *SqliteStore) DeleteCandidate(sessionId int64, candidateId int64) error {
	query := "DELETE FROM " + TableNameCandidates + " WHERE session_id = ? AND id = ?"
	_, err := store.Exec(query, sessionId, candidateId)
	if err != nil {
		return err
	}
	return nil
}

func (store *SqliteStore) AddVote(sessionId int64, vote *core.VoteEntity) (*core.VoteEntity, error) {
	query := "INSERT INTO " + TableNameVotes + " (session_id, user_id, candidate_id) VALUES (?, ?, ?)"
	_, err := store.Exec(query, sessionId, vote.UserId, vote.CandidateId)
	if err != nil {
		return nil, err
	}

	return vote, nil
}

func (store *SqliteStore) GetVotesByUserId(sessionId int64, userId int64) (*[]core.VoteEntity, error) {
	query := "SELECT session_id, user_id, candidate_id FROM " + TableNameVotes + " WHERE session_id = ? AND user_id = ?"
	rows, err := store.db.Query(query, sessionId, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	votes := make([]core.VoteEntity, 0)
	for rows.Next() {
		vote := core.VoteEntity{}
		err := rows.Scan(&vote.SessionId, &vote.UserId, &vote.CandidateId)
		if err != nil {
			return nil, err
		}
		votes = append(votes, vote)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &votes, nil
}

func (store *SqliteStore) GetVote(sessionId int64, userId int64, candidateId int64) (*core.VoteEntity, error) {
	query := "SELECT session_id, user_id, candidate_id FROM " + TableNameVotes + " WHERE session_id = ? AND user_id = ? AND candidate_id = ?"
	row := store.db.QueryRow(query, sessionId, userId, candidateId)
	vote := &core.VoteEntity{}
	err := row.Scan(&vote.SessionId, &vote.UserId, &vote.CandidateId)
	if err == sql.ErrNoRows {
		return nil, nil // Vote not found
	} else if err != nil {
		return nil, err
	}
	return vote, nil
}

func (store *SqliteStore) DeleteVote(sessionId int64, userId int64, candidateId int64) error {
	query := "DELETE FROM " + TableNameVotes + " WHERE session_id = ? AND user_id = ? AND candidate_id = ?"
	_, err := store.Exec(query, sessionId, userId, candidateId)
	if err != nil {
		return err
	}
	return nil
}

func (store *SqliteStore) AddPlaylist(sessionId int64, playlist *core.SessionPlaylistEntity) (*core.SessionPlaylistEntity, error) {
	query := "INSERT INTO " + TableNamePlaylists + " (session_id, user_id, playlist_id) VALUES (?, ?, ?)"
	_, err := store.Exec(query, sessionId, playlist.UserId, playlist.PlaylistId)
	if err != nil {
		return nil, err
	}

	return playlist, nil
}

func (store *SqliteStore) FindPlaylist(sessionId int64, userId int64) (*core.SessionPlaylistEntity, error) {
	query := "SELECT session_id, user_id, playlist_id FROM " + TableNamePlaylists + " WHERE session_id = ? AND user_id = ?"
	row := store.db.QueryRow(query, sessionId, userId)
	playlist := &core.SessionPlaylistEntity{}
	err := row.Scan(&playlist.SessionId, &playlist.UserId, &playlist.PlaylistId)
	if err == sql.ErrNoRows {
		return nil, nil // Playlist not found
	} else if err != nil {
		return nil, err
	}

	return playlist, nil
}
