package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/CaribouBlue/top-spot/internal/entities/session"
	"github.com/CaribouBlue/top-spot/internal/entities/user"

	_ "github.com/mattn/go-sqlite3"
)

type sqliteJsonDb struct {
	dbPath       string
	db           *sql.DB
	UserTable    string
	SessionTable string
}

func NewSqliteJsonDb(dbPath string) (*sqliteJsonDb, error) {
	sqlite := &sqliteJsonDb{
		dbPath:       dbPath,
		UserTable:    "user",
		SessionTable: "session",
	}
	err := sqlite.initDb()
	return sqlite, err
}

func (sqlite *sqliteJsonDb) initDb() error {
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

func (sqlite *sqliteJsonDb) NewCollection(collectionName string) error {
	query := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		data JSONB
	);`, collectionName)

	_, err := sqlite.db.Exec(query)
	return err
}

// TODO: Update this method to automatically set the ID field of the model
func (sqlite *sqliteJsonDb) Insert(tableName string, data []byte) (sql.Result, error) {
	stmt, err := sqlite.db.Prepare(fmt.Sprintf("insert into %s(data) values(?)", tableName))
	if err != nil {
		return nil, err
	}

	defer stmt.Close()
	result, err := stmt.Exec(data)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (sqlite *sqliteJsonDb) SelectAll(tableName string) ([][]byte, error) {
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

func (sqlite *sqliteJsonDb) QueryRow(query string) ([]byte, error) {
	var data []byte
	err := sqlite.db.QueryRow(query).Scan(&data)
	if err != nil {
		return data, err
	}

	return data, err
}

func (sqlite *sqliteJsonDb) SelectOne(tableName string, recordId int64) ([]byte, error) {
	query := fmt.Sprintf("select data from %s where data->>'id' = %d", tableName, recordId)
	return sqlite.QueryRow(query)
}

func (sqlite *sqliteJsonDb) Update(tableName string, recordId any, data []byte) error {
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

func (sqlite *sqliteJsonDb) DeleteRecord(tableName string, recordId any) error {
	return nil
}

func (sqlite *sqliteJsonDb) GetUser(userId int64) (*user.User, error) {
	query := fmt.Sprintf("SELECT data FROM %s WHERE id = %d", sqlite.UserTable, userId)
	log.Default().Println("Query: ", query)
	data, err := sqlite.QueryRow(query)
	if err == sql.ErrNoRows {
		return nil, user.ErrNoUserFound
	} else if err != nil {
		return nil, err
	}

	u := &user.User{}
	err = json.Unmarshal(data, u)
	return u, err
}

func (sqlite *sqliteJsonDb) GetUserByUsername(username string) (*user.User, error) {
	data, err := sqlite.QueryRow(fmt.Sprintf("select data from %s where data->>'userName' = '%s'", sqlite.UserTable, username))
	if err == sql.ErrNoRows {
		return nil, user.ErrNoUserFound
	} else if err != nil {
		return nil, err
	}

	u := &user.User{}
	err = json.Unmarshal(data, u)
	return u, err
}

func (sqlite *sqliteJsonDb) CreateUser(user *user.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	result, err := sqlite.Insert(sqlite.UserTable, data)
	if err != nil {
		return err
	}

	log.Default().Println("User ID: ", user.Id)

	if user.Id == 0 {
		user.Id, err = result.LastInsertId()
		if err != nil {
			return err
		}

		log.Default().Println("New User ID: ", user.Id)

		err = sqlite.UpdateUser(user)
		if err != nil {
			return err
		}
	}

	return nil
}

func (sqlite *sqliteJsonDb) UpdateUser(user *user.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	stmt, err := sqlite.db.Prepare(fmt.Sprintf("UPDATE %s SET data = ? WHERE id = ?", sqlite.UserTable))
	if err != nil {
		return err
	}

	defer stmt.Close()
	_, err = stmt.Exec(data, user.Id)
	if err != nil {
		return err
	}

	return nil
}

func (sqlite *sqliteJsonDb) DeleteUser(user *user.User) error {
	return sqlite.DeleteRecord(sqlite.UserTable, user.Id)
}

func (sqlite *sqliteJsonDb) GetSessions() ([]*session.Session, error) {
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

func (sqlite *sqliteJsonDb) GetSession(sessionId int64) (*session.Session, error) {
	data, err := sqlite.SelectOne(sqlite.SessionTable, sessionId)
	if err != nil {
		return nil, err
	}

	session := &session.Session{}
	err = json.Unmarshal(data, session)
	return session, err
}

func (sqlite *sqliteJsonDb) UpdateSession(session *session.Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	return sqlite.Update(sqlite.SessionTable, session.Id, data)
}

func (sqlite *sqliteJsonDb) CreateSession(session *session.Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	_, err = sqlite.Insert(sqlite.SessionTable, data)
	return err
}
