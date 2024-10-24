package db

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log"
	"path"

	"github.com/CaribouBlue/top-spot/appdata"
	"github.com/CaribouBlue/top-spot/spotify"
	_ "github.com/mattn/go-sqlite3"
)

var _db *sql.DB

func initDb() {
	appDataDir, err := appdata.GetAppDataDir()
	if err != nil {
		log.Fatal(err)
	}

	dbPath := path.Join(appDataDir, "top-spot.db")
	_db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	if err = _db.Ping(); err != nil {
		log.Fatal(err)
	}

	createJsonDataModelTable(NewUserDataModel())
}

func getDb() *sql.DB {
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

	_, err := getDb().Exec(query)
	return err
}

func InsertJsonDataModel(model JsonDataModel) (JsonDataModel, error) {
	stmt, err := getDb().Prepare(fmt.Sprintf("insert into %s(data) values(?)", model.GetTableName()))
	if err != nil {
		return model, err
	}

	value, err := model.Value()
	if err != nil {
		return model, err
	}

	defer stmt.Close()
	result, err := stmt.Exec(value)
	if err != nil {
		return model, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model, err
	}

	model.SetId(id)

	return model, nil
}

func UpdateJsonDataModel(model JsonDataModel) error {
	stmt, err := getDb().Prepare(fmt.Sprintf("update %s set data = ? where data->>'id' = ?", model.GetTableName()))
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

func GetJsonDataModelById(model JsonDataModel) (JsonDataModel, error) {
	var data string
	query := fmt.Sprintf("select data from %s where data->>'id' = %d", model.GetTableName(), model.GetId())
	err := getDb().QueryRow(query).Scan(&data)
	if err != nil {
		return model, err
	}

	err = model.Scan(data)
	return model, err
}

type UserDataModel struct {
	Id                 int64               `json:"id"`
	SpotifyAccessToken spotify.AccessToken `json:"spotifyAccessToken"`
}

func (user *UserDataModel) GetTableName() string {
	return "users"
}

func (user *UserDataModel) SetId(id int64) {
	user.Id = id
}

func (user *UserDataModel) GetId() int64 {
	return user.Id
}

func (user *UserDataModel) Scan(value interface{}) error {
	return json.Unmarshal([]byte(value.(string)), user)
}

func (user *UserDataModel) Value() (driver.Value, error) {
	return json.Marshal(user)
}

func (user *UserDataModel) Insert() error {
	_, err := InsertJsonDataModel(user)
	return err
}

func (user *UserDataModel) Update() error {
	return UpdateJsonDataModel(user)
}

func (user *UserDataModel) GetById() error {
	_, err := GetJsonDataModelById(user)
	return err
}

func (user *UserDataModel) IsAuthenticated() (bool, error) {
	err := user.GetById()
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}

	if user.SpotifyAccessToken.AccessToken == "" {
		return false, nil
	}

	return true, nil
}

func NewUserDataModel() *UserDataModel {
	user := UserDataModel{}
	return &user
}
