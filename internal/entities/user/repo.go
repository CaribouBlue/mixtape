package user

import "errors"

type UserRepo interface {
	GetUser(userId int64) (*User, error)
	GetUserByUsername(username string) (*User, error)
	SearchUsers(query string) (*[]User, error)
	CreateUser(*User) error
	UpdateUser(*User) error
	DeleteUser(*User) error
}

var (
	ErrNoUserFound = errors.New("no user found")
)
