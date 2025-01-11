package db

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/CaribouBlue/top-spot/internal/session"
	"github.com/CaribouBlue/top-spot/internal/user"

	_ "github.com/mattn/go-sqlite3"
)

type sqliteJsonStore struct {
	dbPath       string
	db           *sql.DB
	UserTable    string
	SessionTable string
}

func NewSqliteJsonStore(dbPath string) (*sqliteJsonStore, error) {
	sqlite := &sqliteJsonStore{
		dbPath:       dbPath,
		UserTable:    "user",
		SessionTable: "session",
	}
	err := sqlite.initDb()
	return sqlite, err
}

func (sqlite *sqliteJsonStore) initDb() error {
	db, err := sql.Open("sqlite3", sqlite.dbPath)
	if err != nil {
		return err
	}
	sqlite.db = db

	if err = sqlite.db.Ping(); err != nil {
		return err
	}

	return nil
}

func (sqlite *sqliteJsonStore) NewCollection(collectionName string) error {
	query := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		data jsonb
	);`, collectionName)

	_, err := sqlite.db.Exec(query)
	return err
}

// TODO: Update this method to automatically set the ID field of the model
func (sqlite *sqliteJsonStore) Insert(tableName string, data []byte) error {
	stmt, err := sqlite.db.Prepare(fmt.Sprintf("insert into %s(data) values(?)", tableName))
	if err != nil {
		return err
	}

	defer stmt.Close()
	_, err = stmt.Exec(data)
	if err != nil {
		return err
	}

	return nil
}

func (sqlite *sqliteJsonStore) SelectAll(tableName string) ([][]byte, error) {
	var records [][]byte = make([][]byte, 0)

	query := fmt.Sprintf("select data from %s", tableName)
	rows, err := sqlite.db.Query(query)
	if err != nil {
		return records, err
	}

	defer rows.Close()

	var data []byte
	for rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			return records, err
		}

		records = append(records, data)
	}

	return records, nil
}

func (sqlite *sqliteJsonStore) SelectOne(tableName string, recordId int64) ([]byte, error) {
	var data []byte
	query := fmt.Sprintf("select data from %s where data->>'id' = %d", tableName, recordId)
	err := sqlite.db.QueryRow(query).Scan(&data)
	if err != nil {
		return data, err
	}

	return data, err
}

func (sqlite *sqliteJsonStore) Update(tableName string, recordId any, data []byte) error {
	stmt, err := sqlite.db.Prepare(fmt.Sprintf("update %s set data = ? where data->>'id' = ?", tableName))
	if err != nil {
		return err
	}

	defer stmt.Close()
	_, err = stmt.Exec(data, recordId)
	if err != nil {
		return err
	}

	return nil
}

func (sqlite *sqliteJsonStore) DeleteRecord(tableName string, recordId any) error {
	return nil
}

func (sqlite *sqliteJsonStore) GetUser(userId int64) (*user.User, error) {
	data, err := sqlite.SelectOne(sqlite.UserTable, userId)
	if err == sql.ErrNoRows {
		return nil, user.ErrNoUserFound
	} else if err != nil {
		return nil, err
	}

	u := &user.User{}
	err = json.Unmarshal(data, u)
	return u, err
}

func (sqlite *sqliteJsonStore) CreateUser(user *user.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return sqlite.Insert(sqlite.UserTable, data)
}

func (sqlite *sqliteJsonStore) UpdateUser(user *user.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return sqlite.Update(sqlite.UserTable, user.Id, data)
}

func (sqlite *sqliteJsonStore) DeleteUser(user *user.User) error {
	return sqlite.DeleteRecord(sqlite.UserTable, user.Id)
}

func (sqlite *sqliteJsonStore) GetSessions() ([]*session.Session, error) {
	records, err := sqlite.SelectAll(sqlite.SessionTable)
	if err != nil {
		return nil, err
	}

	sessions := make([]*session.Session, 0)
	for _, record := range records {
		session := &session.Session{}
		err = json.Unmarshal([]byte(record), session)
		if err != nil {
			return nil, err
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (sqlite *sqliteJsonStore) GetSession(sessionId int64) (*session.Session, error) {
	data, err := sqlite.SelectOne(sqlite.SessionTable, sessionId)
	if err != nil {
		return nil, err
	}

	session := &session.Session{}
	err = json.Unmarshal(data, session)
	return session, err
}

func (sqlite *sqliteJsonStore) UpdateSession(session *session.Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	return sqlite.Update(sqlite.SessionTable, session.Id, data)
}

func (sqlite *sqliteJsonStore) CreateSession(session *session.Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	return sqlite.Insert(sqlite.SessionTable, data)
}
