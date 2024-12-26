package db

import (
	"database/sql/driver"
)

const GlobalDbName = "top-spot.db"

var db Db

func InitGlobal() {
	db = NewSqliteJsonDb(GlobalDbName)
}

func Global() Db {
	if db == nil {
		InitGlobal()
	}
	return db
}

type Model interface {
	Name() string
	Id() int64
	Scan(value interface{}) error
	Value() (driver.Value, error)
}

type Db interface {
	Close() error
	NewCollection(Model) error
	CreateRecord(Model) error
	UpdateRecord(Model) error
	DeleteRecord(Model) error
	ReadRecord(Model) error
	ReadRecords(Model) ([]interface{}, error)
}
