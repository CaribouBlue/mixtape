package core

import (
	"errors"
	"os"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrPasswordsDoNotMatch   = errors.New("passwords do not match")
	ErrIncorrectPassword     = errors.New("incorrect password")
	ErrUserNotFound          = errors.New("user not found")
	ErrIncorrectAccessCode   = errors.New("incorrect access code")
)

type UserEntity struct {
	Id             int64  `json:"id"`
	Username       string `json:"username"`
	DisplayName    string `json:"displayName"`
	SpotifyToken   string
	HashedPassword []byte
}

func (u *UserEntity) IdString() string {
	return strconv.FormatInt(u.Id, 10)
}

type UserRepository interface {
	CreateUser(user *UserEntity) (*UserEntity, error)
	GetUserById(userId int64) (*UserEntity, error)
	GetUserByUsername(username string) (*UserEntity, error)
	GetAllUsers() (*[]UserEntity, error)
	UpdateUserSpotifyToken(userId int64, spotifyToken string) (*UserEntity, error)
}

type UserService struct {
	userRepository UserRepository
}

func NewUserService(userRepository UserRepository) *UserService {
	return &UserService{
		userRepository: userRepository,
	}
}

func (s *UserService) NormalizeUsername(username string) string {
	re := regexp.MustCompile(`\s+`)
	normalizedUsername := strings.ToLower(re.ReplaceAllString(username, ""))
	return normalizedUsername
}

func (s *UserService) SignUpNewUser(username, password, confirmPassword, accessCode string) (*UserEntity, error) {
	if accessCode != os.Getenv("ACCESS_CODE") {
		return nil, ErrIncorrectAccessCode
	}

	if password != confirmPassword {
		return nil, ErrPasswordsDoNotMatch
	}

	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &UserEntity{
		Username:       s.NormalizeUsername(username),
		DisplayName:    username,
		HashedPassword: hashedPassword,
	}

	existingUser, err := s.userRepository.GetUserByUsername(user.Username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUsernameAlreadyExists
	}

	return s.userRepository.CreateUser(user)
}

func (s *UserService) LoginUser(username string, password string) (*UserEntity, error) {
	normalizedUsername := s.NormalizeUsername(username)
	user, err := s.userRepository.GetUserByUsername(normalizedUsername)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(password)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return nil, ErrIncorrectPassword
		}

		return nil, err
	}

	return user, nil
}

func (s *UserService) AuthenticateSpotify(userId int64, spotifyToken string) (*UserEntity, error) {
	user, err := s.userRepository.UpdateUserSpotifyToken(userId, spotifyToken)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) IsAuthenticated(user *UserEntity) (bool, error) {
	return user != nil && user.SpotifyToken != "", nil
}

func (s *UserService) GetUserById(userId int64) (*UserEntity, error) {
	user, err := s.userRepository.GetUserById(userId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func HashPassword(password string) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return hash, nil
}
