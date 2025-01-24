package user

type UserService interface {
	Get(userId int64) (*User, error)
	Create(*User) error
	Update(*User) error
	Delete(*User) error

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

func (s *userService) IsAuthenticated(user *User) (bool, error) {
	if user.SpotifyAccessToken.AccessToken != "" {
		return true, nil
	}
	return false, nil
}
