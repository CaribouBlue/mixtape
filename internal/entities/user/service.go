package user

import (
	"errors"
	"log"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUsernameExists    = errors.New("username already exists")
	ErrIncorrectPassword = errors.New("incorrect password")
)

type UserService interface {
	Get(userId int64) (*User, error)
	Create(*User) error
	Update(*User) error
	Delete(*User) error

	SignUp(username string, password string) (*User, error)
	Login(username string, password string) (*User, error)

	IsAuthenticated(*User) (bool, error)
}

type userService struct {
	repo UserRepo
}

func NewUserService(repo UserRepo) UserService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) Get(userId int64) (*User, error) {
	return s.repo.GetUser(userId)
}

func (s *userService) Create(user *User) error {
	return s.repo.CreateUser(user)
}

func (s *userService) Update(user *User) error {
	return s.repo.UpdateUser(user)
}

func (s *userService) Delete(user *User) error {
	return s.repo.DeleteUser(user)
}

func (s *userService) SignUp(username string, password string) (*User, error) {
	log.Default().Println("signing up user")

	if _, err := s.repo.GetUserByUsername(username); err == nil {
		return nil, ErrUsernameExists
	} else if err != ErrNoUserFound {
		return nil, err
	}

	log.Default().Println("creating user")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		Id:           0,
		UserName:     username,
		PasswordHash: hashedPassword,
	}

	err = s.Create(user)
	return user, err
}

func (s *userService) Login(username string, password string) (*User, error) {
	log.Default().Println("logging in user")

	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	log.Default().Println("got user: ", user.Id, user.UserName)

	log.Default().Println("comparing password")

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return nil, ErrIncorrectPassword
		}

		return nil, err
	}

	return user, nil
}

func (s *userService) IsAuthenticated(user *User) (bool, error) {
	if user.SpotifyAccessToken.AccessToken != "" {
		return true, nil
	}
	return false, nil
}
