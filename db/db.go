package db

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log"
	"path"

	"github.com/CaribouBlue/top-spot/appdata"
	_ "github.com/mattn/go-sqlite3"
)

const DbName = "top-spot.db"

var _db *sql.DB

func initDb() {
	appDataDir, err := appdata.GetAppDataDir()
	if err != nil {
		log.Fatal(err)
	}

	dbPath := path.Join(appDataDir, DbName)
	_db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	if err = _db.Ping(); err != nil {
		log.Fatal(err)
	}

	createJsonDataModelTable(NewUserDataModel())
	createJsonDataModelTable(NewGameSessionDataModel())
}

func GetDb() *sql.DB {
	if _db == nil {
		initDb()
	}
	return _db
}

type JsonDataModel interface {
	Scan(value interface{}) error
	Value() (driver.Value, error)
	GetTableName() string
	SetId(id int64)
	GetId() int64
}

func createJsonDataModelTable(model JsonDataModel) error {
	query := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		data jsonb
	);`, model.GetTableName())

	_, err := GetDb().Exec(query)
	return err
}

func insertJsonDataModel(model JsonDataModel) (JsonDataModel, error) {
	stmt, err := GetDb().Prepare(fmt.Sprintf("insert into %s(data) values(?)", model.GetTableName()))
	if err != nil {
		return model, err
	}

	value, err := model.Value()
	if err != nil {
		return model, err
	}

	defer stmt.Close()
	_, err = stmt.Exec(value)
	if err != nil {
		return model, err
	}

	// id, err := result.LastInsertId()
	// if err != nil {
	// 	return model, err
	// }

	// model.SetId(id)

	return model, nil
}

func updateJsonDataModel(model JsonDataModel) error {
	stmt, err := GetDb().Prepare(fmt.Sprintf("update %s set data = ? where data->>'id' = ?", model.GetTableName()))
	if err != nil {
		return err
	}
	userValue, err := model.Value()
	if err != nil {
		return err
	}

	defer stmt.Close()
	_, err = stmt.Exec(userValue, model.GetId())
	if err != nil {
		return err
	}

	return nil
}

func getJsonDataModelById(model JsonDataModel) (JsonDataModel, error) {
	var data string
	query := fmt.Sprintf("select data from %s where data->>'id' = %d", model.GetTableName(), model.GetId())
	err := GetDb().QueryRow(query).Scan(&data)
	if err != nil {
		return model, err
	}

	err = model.Scan(data)
	return model, err
}
