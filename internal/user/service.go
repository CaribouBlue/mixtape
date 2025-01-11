package user

type UserService interface {
	Get(userId int64) (*User, error)
	Create(*User) error
	Update(*User) error
	Delete(*User) error

	IsAuthenticated(*User) (bool, error)
}

type userService struct {
	store UserStore
}

func NewUserService(store UserStore) UserService {
	return &userService{
		store: store,
	}
}

func (s *userService) Get(userId int64) (*User, error) {
	return s.store.GetUser(userId)
}

func (s *userService) Create(user *User) error {
	return s.store.CreateUser(user)
}

func (s *userService) Update(user *User) error {
	return s.store.UpdateUser(user)
}

func (s *userService) Delete(user *User) error {
	return s.store.DeleteUser(user)
}

func (s *userService) IsAuthenticated(user *User) (bool, error) {
	if user.SpotifyAccessToken.AccessToken != "" {
		return true, nil
	}
	return false, nil
}
