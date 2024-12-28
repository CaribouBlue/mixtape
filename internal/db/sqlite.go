package db

import (
	"database/sql"
	"fmt"
	"log"
	"path"

	"github.com/CaribouBlue/top-spot/internal/appdata"
	_ "github.com/mattn/go-sqlite3"
)

type SqliteJsonDb struct {
	name string
	db   *sql.DB
}

func NewSqliteJsonDb(name string) *SqliteJsonDb {
	sqlite := &SqliteJsonDb{
		name: name,
	}
	sqlite.Db()
	return sqlite
}

func (sqlite *SqliteJsonDb) Db() *sql.DB {
	if sqlite.db == nil {
		appDataDir, err := appdata.GetAppDataDir()
		if err != nil {
			log.Fatal(err)
		}

		dbPath := path.Join(appDataDir, sqlite.name)
		sqlite.db, err = sql.Open("sqlite3", dbPath)
		if err != nil {
			log.Fatal(err)
		}

		if err = sqlite.db.Ping(); err != nil {
			log.Fatal(err)
		}
	}
	return sqlite.db
}

func (sqlite *SqliteJsonDb) Close() error {
	return sqlite.Db().Close()
}

func (sqlite *SqliteJsonDb) NewCollection(model Model) error {
	query := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		data jsonb
	);`, model.Name())

	_, err := sqlite.Db().Exec(query)
	return err
}

// TODO: Update this method to automatically set the ID field of the model
func (sqlite *SqliteJsonDb) CreateRecord(model Model) error {
	stmt, err := sqlite.Db().Prepare(fmt.Sprintf("insert into %s(data) values(?)", model.Name()))
	if err != nil {
		return err
	}

	value, err := model.Value()
	if err != nil {
		return err
	}

	defer stmt.Close()
	_, err = stmt.Exec(value)
	if err != nil {
		return err
	}

	return nil
}

func (sqlite *SqliteJsonDb) ReadRecords(model Model) ([]interface{}, error) {
	var records []interface{} = make([]interface{}, 0)

	query := fmt.Sprintf("select data from %s", model.Name())
	rows, err := sqlite.Db().Query(query)
	if err != nil {
		return records, err
	}

	defer rows.Close()

	var data string
	for rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			return records, err
		}

		records = append(records, data)
	}

	return records, nil
}

func (sqlite *SqliteJsonDb) ReadRecord(model Model) error {
	var data string
	query := fmt.Sprintf("select data from %s where data->>'id' = %d", model.Name(), model.Id())
	err := sqlite.Db().QueryRow(query).Scan(&data)
	if err != nil {
		return err
	}

	err = model.Scan(data)
	return err
}

func (sqlite *SqliteJsonDb) UpdateRecord(model Model) error {
	stmt, err := sqlite.Db().Prepare(fmt.Sprintf("update %s set data = ? where data->>'id' = ?", model.Name()))
	if err != nil {
		return err
	}
	userValue, err := model.Value()
	if err != nil {
		return err
	}

	defer stmt.Close()
	_, err = stmt.Exec(userValue, model.Id())
	if err != nil {
		return err
	}

	return nil
}

func (sqlite *SqliteJsonDb) DeleteRecord(model Model) error {
	return nil
}
