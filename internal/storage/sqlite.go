package storage

import (
	"database/sql"
	"time"

	"github.com/CaribouBlue/top-spot/internal/core"
	_ "github.com/mattn/go-sqlite3"
)

const (
	TableNameUsers       = "users"
	TableNameSessions    = "sessions"
	TableNameSubmissions = "submissions"
	TableNameVotes       = "votes"
	TableNamePlaylists   = "playlists"
	ViewNameCandidates   = "candidates"
)

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

func (store *SqliteStore) UpdateSession(session *core.SessionEntity) (*core.SessionEntity, error) {
	query := `UPDATE ` + TableNameSessions + `
		SET name = ?, 
			created_by = ?,
			created_at = ?,
			max_submissions = ?,
			start_at = ?,
			submission_phase_duration = ?,
			vote_phase_duration = ?
		WHERE id = ?`
	_, err := store.Exec(query, session.Name, session.CreatedBy, session.CreatedAt.Unix(), session.MaxSubmissions, session.StartAt.Unix(), session.SubmissionPhaseDuration, session.VotePhaseDuration, session.Id)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (store *SqliteStore) DeleteSession(id int64) error {
	query := "DELETE FROM " + TableNameSessions + " WHERE id = ?"
	_, err := store.Exec(query, id)
	if err != nil {
		return err
	}
	return nil
}

func (store *SqliteStore) AddSubmission(sessionId int64, submission *core.SubmissionEntity) (*core.SubmissionEntity, error) {
	query := "INSERT INTO " + TableNameSubmissions + " (session_id, user_id, track_id) VALUES (?, ?, ?)"
	result, err := store.Exec(query, sessionId, submission.UserId, submission.TrackId)
	if err != nil {
		return nil, err
	}

	submissionId, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	submission.Id = submissionId

	return submission, nil
}

func (store *SqliteStore) GetAllSubmissions(sessionId int64) (*[]core.SubmissionEntity, error) {
	query := "SELECT id, session_id, user_id, track_id FROM " + TableNameSubmissions + " WHERE session_id = ?"
	rows, err := store.db.Query(query, sessionId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	submissions := make([]core.SubmissionEntity, 0)
	for rows.Next() {
		submission := core.SubmissionEntity{}
		err := rows.Scan(&submission.Id, &submission.SessionId, &submission.UserId, &submission.TrackId)
		if err != nil {
			return nil, err
		}
		submissions = append(submissions, submission)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &submissions, nil
}

func (store *SqliteStore) GetSubmissionById(sessionId int64, submissionId int64) (*core.SubmissionEntity, error) {
	query := "SELECT id, session_id, user_id, track_id FROM " + TableNameSubmissions + " WHERE session_id = ? AND id = ?"
	row := store.db.QueryRow(query, sessionId, submissionId)
	submission := &core.SubmissionEntity{}
	err := row.Scan(&submission.Id, &submission.SessionId, &submission.UserId, &submission.TrackId)
	if err == sql.ErrNoRows {
		return nil, nil // Submission not found
	} else if err != nil {
		return nil, err
	}
	return submission, nil
}

func (store *SqliteStore) GetSubmissionsByUserId(sessionId int64, userId int64) (*[]core.SubmissionEntity, error) {
	query := "SELECT id, session_id, user_id, track_id FROM " + TableNameSubmissions + " WHERE session_id = ? AND user_id = ?"
	rows, err := store.db.Query(query, sessionId, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	submissions := make([]core.SubmissionEntity, 0)
	for rows.Next() {
		submission := core.SubmissionEntity{}
		err := rows.Scan(&submission.Id, &submission.SessionId, &submission.UserId, &submission.TrackId)
		if err != nil {
			return nil, err
		}
		submissions = append(submissions, submission)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &submissions, nil
}

func (store *SqliteStore) DeleteSubmission(sessionId int64, submissionId int64) error {
	query := "DELETE FROM " + TableNameSubmissions + " WHERE session_id = ? AND id = ?"
	_, err := store.Exec(query, sessionId, submissionId)
	if err != nil {
		return err
	}
	return nil
}

func (store *SqliteStore) AddVote(sessionId int64, vote *core.VoteEntity) (*core.VoteEntity, error) {
	query := "INSERT INTO " + TableNameVotes + " (session_id, user_id, submission_id) VALUES (?, ?, ?)"
	_, err := store.Exec(query, sessionId, vote.UserId, vote.SubmissionId)
	if err != nil {
		return nil, err
	}

	return vote, nil
}

func (store *SqliteStore) GetAllVotes(sessionId int64) (*[]core.VoteEntity, error) {
	query := "SELECT session_id, user_id, submission_id FROM " + TableNameVotes + " WHERE session_id = ?"
	rows, err := store.db.Query(query, sessionId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	votes := make([]core.VoteEntity, 0)
	for rows.Next() {
		vote := core.VoteEntity{}
		err := rows.Scan(&vote.SessionId, &vote.UserId, &vote.SubmissionId)
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

func (store *SqliteStore) GetVotesByUserId(sessionId int64, userId int64) (*[]core.VoteEntity, error) {
	query := "SELECT session_id, user_id, submission_id FROM " + TableNameVotes + " WHERE session_id = ? AND user_id = ?"
	rows, err := store.db.Query(query, sessionId, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	votes := make([]core.VoteEntity, 0)
	for rows.Next() {
		vote := core.VoteEntity{}
		err := rows.Scan(&vote.SessionId, &vote.UserId, &vote.SubmissionId)
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

func (store *SqliteStore) DeleteVote(sessionId int64, userId int64, submissionId int64) error {
	query := "DELETE FROM " + TableNameVotes + " WHERE session_id = ? AND user_id = ? AND submission_id = ?"
	_, err := store.Exec(query, sessionId, userId, submissionId)
	if err != nil {
		return err
	}
	return nil
}

func (store *SqliteStore) GetUserCandidates(sessionId int64, userId int64) (*[]core.CandidateDto, error) {
	query := `
		SELECT submission_id,
			submission_user_id,
			track_id,
			vote_submission_id,
			vote_user_id
		FROM ` + ViewNameCandidates + `
		WHERE session_id = ? 
			AND vote_user_id = ?;
	`
	rows, err := store.db.Query(query, sessionId, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	candidates := make([]core.CandidateDto, 0)
	for rows.Next() {
		candidate := core.NewCandidateDto(sessionId)
		var voteSubmissionId *int64
		err := rows.Scan(&candidate.Submission.Id, &candidate.Submission.UserId, &candidate.Submission.TrackId, &voteSubmissionId, &candidate.Vote.UserId)
		if err != nil {
			return nil, err
		}

		if voteSubmissionId == nil {
			candidate.Vote = nil
		} else {
			candidate.Vote.SubmissionId = *voteSubmissionId
		}

		candidates = append(candidates, *candidate)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &candidates, nil
}

func (store *SqliteStore) GetCandidate(sessionId int64, userId int64, submissionId int64) (*core.CandidateDto, error) {
	query := `
		SELECT submission_id,
			submission_user_id,
			track_id,
			vote_submission_id,
			vote_user_id
		FROM ` + ViewNameCandidates + `
		WHERE session_id = ? 
			AND vote_user_id = ?
			AND submission_id = ?;
	`

	row := store.db.QueryRow(query, sessionId, userId, submissionId)
	candidate := core.NewCandidateDto(sessionId)
	var voteSubmissionId *int64
	err := row.Scan(&candidate.Submission.Id, &candidate.Submission.UserId, &candidate.Submission.TrackId, &voteSubmissionId, &candidate.Vote.UserId)
	if err == sql.ErrNoRows {
		return nil, nil // Candidate not found
	} else if err != nil {
		return nil, err
	}

	if voteSubmissionId == nil {
		candidate.Vote = nil
	} else {
		candidate.Vote.SubmissionId = *voteSubmissionId
	}

	return candidate, nil
}

func (store *SqliteStore) GetCandidatesWithVotes(sessionId int64) (*[]core.CandidateDto, error) {
	query := `
		SELECT submission_id,
			submission_user_id,
			track_id,
			vote_submission_id,
			vote_user_id
		FROM candidates
		WHERE session_id = ?
			AND vote_submission_id IS NOT NULL
		ORDER BY submission_id;
	`
	rows, err := store.db.Query(query, sessionId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	candidates := make([]core.CandidateDto, 0)
	for rows.Next() {
		candidate := core.NewCandidateDto(sessionId)
		err := rows.Scan(&candidate.Submission.Id, &candidate.Submission.UserId, &candidate.Submission.TrackId, &candidate.Vote.SubmissionId, &candidate.Vote.UserId)
		if err != nil {
			return nil, err
		}

		candidates = append(candidates, *candidate)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &candidates, nil
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

func (store *SqliteStore) DeletePlaylist(sessionId int64, userId int64) error {
	query := "DELETE FROM " + TableNamePlaylists + " WHERE session_id = ? AND user_id = ?"
	_, err := store.Exec(query, sessionId, userId)
	if err != nil {
		return err
	}
	return nil
}
